// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestTemplateRulesWithNonTypedArrayFields(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml := `
---
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
      initContainers: null
      containers:
      - name: simple-app
        image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
        env:
        - name: HELLO_MSG
          value: stranger
        envFrom:
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: simple-cm
  annotations:
    kapp.k14s.io/versioned: ""
data:
  hello_msg: carvel
`

	name := "test-templaterules-with-non-typed-array-fields"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()
	logger.Section("Initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-"}, RunOpts{StdinReader: strings.NewReader(yaml)})

		NewPresentClusterResource("deployment", "simple-app", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "simple-cm-ver-1", env.Namespace, kubectl)
	})
}
