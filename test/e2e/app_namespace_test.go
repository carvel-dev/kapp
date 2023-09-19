// Copyright 2023 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppNamespace(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-app-namespace"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	yaml := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
`

	logger.Section("deploy app with -n flag", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-n", env.Namespace}, RunOpts{NoNamespace: true, StdinReader: strings.NewReader(yaml)})

		// both app meta configmap and the resources should be present in the -n namespace
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "test-cm", env.Namespace, kubectl)
	})

	logger.Section("deploy same app with both -n and --app-namespace flag", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-n", env.Namespace, "--app-namespace", env.Namespace, "--diff-run", "--diff-exit-status"},
			RunOpts{NoNamespace: true, AllowError: true, StdinReader: strings.NewReader(yaml)})

		require.Errorf(t, err, "Expected to receive error")

		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("deploy same app with --app-namespace flag only", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--app-namespace", env.Namespace, "--diff-run", "--diff-exit-status"},
			RunOpts{NoNamespace: true, AllowError: true, StdinReader: strings.NewReader(yaml)})

		require.Errorf(t, err, "Expected to receive error")
		// Resources would get created in the default namespace from kubeconfig since -n is not provided
		require.Containsf(t, err.Error(), "Exiting after diffing with pending changes (exit status 3)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '3'", "Expected to find exit code")
	})

	logger.Section("delete app with --app-namespace flag only", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name, "--app-namespace", env.Namespace},
			RunOpts{NoNamespace: true})

		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "test-cm", env.Namespace, kubectl)
	})
}
