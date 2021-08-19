// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestWaitTimeout(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: ns1
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
apiVersion: v1
kind: Namespace
metadata:
  name: ns2
---
apiVersion: v1
kind: Namespace
metadata:
  name: ns3
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: ns2
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
  namespace: ns3
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
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2m", "--wait-resource-timeout", "2ms", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		expected := []map[string]string{{
			"conditions":      "",
			"kind":            "Namespace",
			"name":            "ns1",
			"namespace":       "(cluster)",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Namespace",
			"name":            "ns2",
			"namespace":       "(cluster)",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Namespace",
			"name":            "ns3",
			"namespace":       "(cluster)",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Deployment",
			"name":            "simple-app",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Deployment",
			"name":            "simple-app2",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Deployment",
			"name":            "simple-app3",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expected, "Op:      6 create, 0 delete, 0 update, 0 noop",
			"Wait to: 6 reconcile, 0 delete, 0 noop", out)

		if len(resp.Lines) > 0 && resp.Lines[len(resp.Lines)-1] != "kapp: Error: waiting on reconcile namespace/ns3 (v1) cluster:\n  Errored:\n    Resource timed out waiting after 2ms" {
			t.Fatalf("Expected to see timed out, but did not: '%s'", resp.Lines[len(resp.Lines)-1])
		}
	})

	cleanUp()

	logger.Section("Global wait timeout", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2s", "--wait-resource-timeout", "2m", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		expected := []map[string]string{{
			"conditions":      "",
			"kind":            "Namespace",
			"name":            "ns1",
			"namespace":       "(cluster)",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Namespace",
			"name":            "ns2",
			"namespace":       "(cluster)",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Namespace",
			"name":            "ns3",
			"namespace":       "(cluster)",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Deployment",
			"name":            "simple-app",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Deployment",
			"name":            "simple-app2",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Deployment",
			"name":            "simple-app3",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expected, "Op:      6 create, 0 delete, 0 update, 0 noop",
			"Wait to: 6 reconcile, 0 delete, 0 noop", out)

		if len(resp.Lines) > 0 && resp.Lines[len(resp.Lines)-1] != "kapp: Error: Timed out waiting after 2s" {
			t.Fatalf("Expected to see timed out, but did not: '%s'", resp.Lines[len(resp.Lines)-1])
		}
	})
}
