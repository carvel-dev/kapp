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
}

func BuildEnv(t *testing.T) Env {
	kappPath := os.Getenv("KAPP_BINARY_PATH")
	if kappPath == "" {
		kappPath = "kapp"
	}

	env := Env{
		Namespace:      os.Getenv("KAPP_E2E_NAMESPACE"),
		KappBinaryPath: kappPath,
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
