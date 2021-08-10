// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"strings"
	"testing"
)

func TestNoDeprecationWarnings(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	name := "test-no-warnings"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	yaml1 := `
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: itachi
spec:
  rules:
  - host: localhost
    http:
      paths:
      - backend:
          serviceName: itachi
          servicePort: 80
          path: /
          pathType: ImplementationSpecific
`
	logger.Section("deploying without --no-deprecation-warning flag", func() {
		out := new(bytes.Buffer)
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml1), StderrWriter: out})
		if !strings.Contains(out.String(), "extensions/v1beta1 Ingress is deprecated in v1.14+, unavailable in v1.22+; use networking.k8s.io/v1 Ingress") {
			t.Fatalf("Expected deprecation warnings, but didn't get")
		}
	})
	cleanUp()
	logger.Section("deploying with --no-deprecation-warning flag", func() {
		out := new(bytes.Buffer)
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--no-deprecation-warnings"},
			RunOpts{StdinReader: strings.NewReader(yaml1), StderrWriter: out})
		if strings.Contains(out.String(), "extensions/v1beta1 Ingress is deprecated in v1.14+, unavailable in v1.22+; use networking.k8s.io/v1 Ingress") {
			t.Fatalf("Expected no deprecation warnings, but got")
		}
	})
}
