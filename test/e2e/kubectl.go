// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Kubectl struct {
	t         *testing.T
	namespace string
	l         Logger
}

func (k Kubectl) Run(args []string) string {
	out, _ := k.RunWithOpts(args, RunOpts{})
	return out
}

func (k Kubectl) RunWithOpts(args []string, opts RunOpts) (string, error) {
	if !opts.NoNamespace {
		args = append(args, []string{"-n", k.namespace}...)
	}

	k.l.Debugf("Running '%s'...\n", k.cmdDesc(args, opts))

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command("kubectl", args...)
	cmd.Stderr = &stderr

	if opts.CancelCh != nil {
		go func() {
			select {
			case <-opts.CancelCh:
				cmd.Process.Signal(os.Interrupt)
			}
		}()
	}

	if opts.StdoutWriter != nil {
		cmd.Stdout = opts.StdoutWriter
	} else {
		cmd.Stdout = &stdout
	}

	cmd.Stdin = opts.StdinReader

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("Execution error: stderr: '%s' error: '%s'", stderr.String(), err)

		require.Truef(k.t, opts.AllowError, "Failed to successfully execute '%s': %v", k.cmdDesc(args, opts), err)
	}

	return stdout.String(), err
}

func (k Kubectl) cmdDesc(args []string, opts RunOpts) string {
	prefix := "kubectl"
	if opts.Redact {
		return prefix + " -redacted-"
	}
	return fmt.Sprintf("%s %s", prefix, strings.Join(args, " "))
}
