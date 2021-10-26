// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestVersionedExplicitReference(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-1
  annotations:
    kapp.k14s.io/versioned: ""
data:
  foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-2
  annotations:
    kapp.k14s.io/versioned-explicit-ref: |
      apiVersion: v1
      kind: ConfigMap
      name: config-1
data:
foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-3
  annotations:
    kapp.k14s.io/versioned-explicit-ref.match: |
      apiVersion: v1
      kind: ConfigMap
      name: config-1
data:
  foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-4
  annotations:
    kapp.k14s.io/versioned-explicit-ref.nomatch: |
      apiVersion: v1
      kind: ConfigMap
      name: config-2
data:
  foo: bar
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-1
  annotations:
    kapp.k14s.io/versioned: ""
data:
  foo: alpha
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-2
  annotations:
    kapp.k14s.io/versioned-explicit-ref: |
      apiVersion: v1
      kind: ConfigMap
      name: config-1
data:
  foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-3
  annotations:
    kapp.k14s.io/versioned-explicit-ref.match: |
      apiVersion: v1
      kind: ConfigMap
      name: config-1
data:
  foo: bar
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-4
  annotations:
    kapp.k14s.io/versioned-explicit-ref.nomatch: |
      apiVersion: v1
      kind: ConfigMap
      name: config-2
data:
  foo: bar
`

	name := "test-versioned-explicit-references"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{
			{
				"age":             "",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-1-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
			{
				"age":             "",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-2",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
			{
				"age":             "",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-3",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
			{
				"age":             "",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-4",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
		}

		require.Exactlyf(t, expected, resp.Tables[0].Rows, "Expected to see correct changes")
	})

	logger.Section("update versioned resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{
			{
				"age":             "",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-1-ver-2",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
			{
				"age":             "<replaced>",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-2",
				"namespace":       "kapp-test",
				"op":              "update",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "ok",
				"wait_to":         "reconcile",
			},
			{
				"age":             "<replaced>",
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-3",
				"namespace":       "kapp-test",
				"op":              "update",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "ok",
				"wait_to":         "reconcile",
			},
		}

		require.Exactlyf(t, expected, replaceAge(resp.Tables[0].Rows), "Expected to see correct changes")
	})
}
