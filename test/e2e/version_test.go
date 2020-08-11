// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	out, _ := kapp.RunWithOpts([]string{"version"}, RunOpts{NoNamespace: true})

	if !strings.Contains(out, "kapp version") {
		t.Fatalf("Expected to find client version")
	}
}
