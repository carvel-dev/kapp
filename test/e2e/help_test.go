// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpCommandGroup(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	_, err := kapp.RunWithOpts([]string{"app-group"}, RunOpts{NoNamespace: true, AllowError: true})
	require.Errorf(t, err, "Expected to receive error")

	require.Contains(t, err.Error(), "Error: Use one of available subcommands: delete, deploy", "Expected helpful error message")
}
