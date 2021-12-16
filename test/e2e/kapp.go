// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"fmt"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kapp/pkg/kapp/cmd"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Kapp struct {
	t         *testing.T
	namespace string
	kappPath  string
	l         Logger
}

type RunOpts struct {
	NoNamespace  bool
	IntoNs       bool
	AllowError   bool
	StderrWriter io.Writer
	StdoutWriter io.Writer
	StdinReader  io.Reader
	CancelCh     chan struct{}
	Redact       bool
	Interactive  bool
}

func (k Kapp) Run(args []string) string {
	out, _ := k.RunWithOpts(args, RunOpts{})
	return out
}

func (k Kapp) RunWithOpts(args []string, opts RunOpts) (string, error) {
	if !opts.NoNamespace {
		args = append(args, []string{"-n", k.namespace}...)
	}
	if opts.IntoNs {
		args = append(args, []string{"--into-ns", k.namespace}...)
	}
	if !opts.Interactive {
		args = append(args, "--yes")
	}

	k.l.Debugf("Running '%s'...\n", k.cmdDesc(args, opts))

	cmd := exec.Command(k.kappPath, args...)
	cmd.Stdin = opts.StdinReader

	var stderr, stdout bytes.Buffer

	if opts.StderrWriter != nil {
		cmd.Stderr = opts.StderrWriter
	} else {
		cmd.Stderr = &stderr
	}

	if opts.StdoutWriter != nil {
		cmd.Stdout = opts.StdoutWriter
	} else {
		cmd.Stdout = &stdout
	}

	if opts.CancelCh != nil {
		go func() {
			select {
			case <-opts.CancelCh:
				cmd.Process.Signal(os.Interrupt)
			}
		}()
	}

	err := cmd.Run()
	stdoutStr := stdout.String()

	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		err = fmt.Errorf("Execution error: stdout: '%s' stderr: '%s' error: '%s' exit code: '%d'",
			stdoutStr, stderr.String(), err, exitCode)

		require.Truef(k.t, opts.AllowError, "Failed to successfully execute '%s': %v", k.cmdDesc(args, opts), err)
	}

	return stdoutStr, err
}

func (k Kapp) RunEmbedded(args []string, opts RunOpts) (string, error) {
	var stdoutBuf bytes.Buffer
	//var stdout io.Writer = bufio.NewWriter(&stdoutBuf)
	var stdout io.Writer = &stdoutBuf

	if opts.StdoutWriter != nil {
		stdout = opts.StdoutWriter
	}

	confUI := ui.NewWrappingConfUI(ui.NewWriterUI(stdout, os.Stderr, ui.NewNoopLogger()), ui.NewNoopLogger())
	defer confUI.Flush()

	if !opts.NoNamespace {
		args = append(args, []string{"-n", k.namespace}...)
	}
	if opts.IntoNs {
		args = append(args, []string{"--into-ns", k.namespace}...)
	}
	if !opts.Interactive {
		args = append(args, "--yes")
	}

	if opts.StdinReader != nil {
		stdin, err := ioutil.ReadAll(opts.StdinReader)
		if err != nil {
			return "", fmt.Errorf("stdin err: %s", err)
		}
		tmpFile, err := newTmpFileSimple(string(stdin))
		if err != nil {
			return "", fmt.Errorf("tmpfile err: %s", err)
		}
		defer os.Remove(tmpFile.Name())
		replaceArg(args, "-", tmpFile.Name())
	}

	command := cmd.NewDefaultKappCmd(confUI)
	command.SetArgs(args)

	err := command.Execute()
	confUI.Flush()

	if err != nil {
		require.Truef(k.t, opts.AllowError, "Failed to successfully execute '%s': %v", k.cmdDesc(args, opts), err)
	}
	return stdoutBuf.String(), err
}

func replaceArg(s []string, elem, replacement string) {
	for i, x := range s {
		if x == elem {
			s[i] = replacement
		}
	}
}

func (k Kapp) cmdDesc(args []string, opts RunOpts) string {
	prefix := "kapp"
	if opts.Redact {
		return prefix + " -redacted-"
	}
	return fmt.Sprintf("%s %s", prefix, strings.Join(args, " "))
}

func newTmpFileSimple(content string) (*os.File, error) {
	file, err := ioutil.TempFile("", "kapp-e2e")
	if err != nil {
		return nil, err
	}

	_, err = file.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return file, nil
}
