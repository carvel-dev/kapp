// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateRetryOnConflict_WithoutConflict(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
  selector:
    app: redis
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6381
  selector:
    app: redis
`

	yamlBehindScenesChange := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6383
  selector:
    app: redis
`

	yamlBehindScenesChangeUndo := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
  selector:
    app: redis
`

	name := "test-retry-on-conflict-without-conflict"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy service successfully even if it was changed but diff remains same", func() {
		promptOutput := newPromptOutput(t)

		go func() {
			promptOutput.WaitPresented()

			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
				RunOpts{IntoNs: true, StdinReader: strings.NewReader(yamlBehindScenesChange)})

			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
				RunOpts{IntoNs: true, StdinReader: strings.NewReader(yamlBehindScenesChangeUndo)})

			promptOutput.WriteYes()
		}()

		tmpFile := newTmpFile(yaml2, t)
		defer os.Remove(tmpFile.Name())

		kapp.RunWithOpts([]string{"deploy", "--tty", "-f", tmpFile.Name(), "-a", name},
			RunOpts{IntoNs: true, StdinReader: promptOutput.YesReader(),
				StdoutWriter: promptOutput.OutputWriter(), Interactive: true})
	})
}

func TestUpdateRetryOnConflict_WithConflict(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
  selector:
    app: redis
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6381
  selector:
    app: redis
`

	yamlBehindScenesChange := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
  selector:
    app: redis
    changed: label
`

	name := "test-retry-on-conflict-without-conflict"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy service and fail because it was changed and diff is different", func() {
		promptOutput := newPromptOutput(t)

		go func() {
			promptOutput.WaitPresented()

			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
				RunOpts{IntoNs: true, StdinReader: strings.NewReader(yamlBehindScenesChange)})

			promptOutput.WriteYes()
		}()

		tmpFile := newTmpFile(yaml2, t)
		defer os.Remove(tmpFile.Name())

		_, err := kapp.RunWithOpts([]string{"deploy", "--tty", "-f", tmpFile.Name(), "-a", name},
			RunOpts{IntoNs: true, StdinReader: promptOutput.YesReader(),
				StdoutWriter: promptOutput.OutputWriter(), Interactive: true, AllowError: true})
		require.Errorf(t, err, "Expected error, but err was nil")

		require.Contains(t, err.Error(), "Failed to update due to resource conflict (approved diff no longer matches)",
			"Expected error to include resource conflict description")

		require.Contains(t, err.Error(), "please apply your changes to the latest version and try again (reason: Conflict)",
			"Expected error to include k8s reason")
	})
}

func TestUpdateRetryOnConflict_WithConflictRebasedAway(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
  selector:
    app: redis
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6381
  selector:
    app: redis
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
rebaseRules:
- path: [spec, selector]
  type: copy
  sources: [existing]
  resourceMatchers:
  - allMatcher: {}
`

	yamlBehindScenesChange := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
  selector:
    app: redis
    changed: label
`

	name := "test-retry-on-conflict-without-conflict"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy service successfully after rebasing and seeing that diff is the same", func() {
		promptOutput := newPromptOutput(t)

		go func() {
			promptOutput.WaitPresented()

			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
				RunOpts{IntoNs: true, StdinReader: strings.NewReader(yamlBehindScenesChange)})

			promptOutput.WriteYes()
		}()

		tmpFile := newTmpFile(yaml2, t)
		defer os.Remove(tmpFile.Name())

		kapp.RunWithOpts([]string{"deploy", "--tty", "-f", tmpFile.Name(), "-a", name},
			RunOpts{IntoNs: true, StdinReader: promptOutput.YesReader(),
				StdoutWriter: promptOutput.OutputWriter(), Interactive: true})
	})
}

type promptOutput struct {
	t            *testing.T
	yesWriter    io.Writer
	yesReader    io.Reader
	outputWriter io.Writer
	outputReader io.Reader
}

func newPromptOutput(t *testing.T) promptOutput {
	yesReader, yesWriter, err := os.Pipe()
	require.NoError(t, err)

	outputReader, outputWriter, err := os.Pipe()
	require.NoError(t, err)

	return promptOutput{t, yesWriter, yesReader, outputWriter, outputReader}
}

func (p promptOutput) WriteYes()            { p.yesWriter.Write([]byte("y\n")) }
func (p promptOutput) YesReader() io.Reader { return p.yesReader }

func (p promptOutput) OutputWriter() io.Writer { return p.outputWriter }
func (p promptOutput) WaitPresented() {
	scanner := bufio.NewScanner(p.outputReader)
	for scanner.Scan() {
		// Cannot easily wait for prompt as it's not NL terminated
		if strings.HasPrefix(scanner.Text(), "Wait to:") {
			break
		}
	}
	err := scanner.Err()
	require.NoError(p.t, err)
}

func newTmpFile(content string, t *testing.T) *os.File {
	file, err := ioutil.TempFile("", "kapp-test-update-retry-on-conflict")
	require.NoError(t, err)

	_, err = file.Write([]byte(content))
	require.NoError(t, err)

	err = file.Close()
	require.NoError(t, err)

	return file
}
