// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestWaitTimeout(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-app
spec:
  selector:
    matchLabels:
      simple-app: ""
  template:
    metadata:
      labels:
        simple-app: ""
    spec:
      containers:
        - name: simple-app
          image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
          env:
            - name: HELLO_MSG
              value: stranger
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-app2
spec:
  selector:
    matchLabels:
      simple-app2: ""
  template:
    metadata:
      labels:
        simple-app2: ""
    spec:
      containers:
        - name: simple-app2
          image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
          env:
            - name: HELLO_MSG
              value: stranger
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-app3
spec:
  selector:
    matchLabels:
      simple-app3: ""
  template:
    metadata:
      labels:
        simple-app3: ""
    spec:
      containers:
        - name: simple-app3
          image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
          env:
            - name: HELLO_MSG
              value: stranger
`

	name := "test-wait-timeout"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Resource wait timeout", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2m", "--wait-resource-timeout", "2ms", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if !strings.Contains(err.Error(), "Resource timed out waiting after 2ms") {
			t.Fatalf("Expected to see timed out, but did not: '%s'", err.Error())
		}
	})

	cleanUp()

	logger.Section("Global wait timeout", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2ms", "--wait-resource-timeout", "2m", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if !strings.Contains(err.Error(), "kapp: Error: Timed out waiting after 2ms") {
			t.Fatalf("Expected to see timed out, but did not: '%s'", err.Error())
		}
	})
}
