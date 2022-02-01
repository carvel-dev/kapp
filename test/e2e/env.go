// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Env struct {
	Namespace      string
	KappBinaryPath string
	SSAEnabled     bool
}

type TestOption struct {
	ServerSideSkip bool
}

// Skip this test when server-side testing mode enabled
// Tests testing rebaseRules should be skipped as rebase
// is completely bypassed in server side mode
func SSASkip(o *TestOption) {
	o.ServerSideSkip = true
}

type TestOptionfunc func(*TestOption)

func BuildEnv(t *testing.T, optFunc ...TestOptionfunc) Env {
	to := TestOption{}
	for _, f := range optFunc {
		f(&to)
	}

	kappPath := os.Getenv("KAPP_BINARY_PATH")
	if kappPath == "" {
		kappPath = "kapp"
	}

	env := Env{
		Namespace:      os.Getenv("KAPP_E2E_NAMESPACE"),
		KappBinaryPath: kappPath,
	}

	if os.Getenv("KAPP_E2E_SSA") == "1" {
		if to.ServerSideSkip {
			t.Skip("SSA incompatible test")
		}
		env.SSAEnabled = true
	}

	env.Validate(t)
	return env
}

func (e Env) Validate(t *testing.T) {
	errStrs := []string{}

	if len(e.Namespace) == 0 {
		errStrs = append(errStrs, "Expected Namespace to be non-empty")
	}

	require.Lenf(t, errStrs, 0, "%s", strings.Join(errStrs, "\n"))

}
