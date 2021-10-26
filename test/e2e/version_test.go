// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	out, _ := kapp.RunWithOpts([]string{"version"}, RunOpts{NoNamespace: true})

	require.Contains(t, out, "kapp version", "Expected to find client version")
}
