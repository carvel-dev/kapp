// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestDiffFilter(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	serviceResourceYaml := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
  namespace: kapp-test
  labels:
    x: "y"
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
---
`
	configMapResourceyYaml := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: kapp-test
  labels:
    x: "z"
data:
  key: value
---
`

	modifiedServiceResourceYaml := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
  namespace: kapp-test
  labels:
    x: "y"
spec:
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: redis
    tier: backend
---
`

	modifiedConfigMapResourceyYaml := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: kapp-test
  labels:
    x: "z"
data:
  key: value2
---
`

	name := "test-diff-filter"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()
	logger.Section("diff filter by label on new resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--diff-filter", `{"newResource": {"labels": ["x=z"]}}`, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(serviceResourceYaml + configMapResourceyYaml)})

		expectedChange := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("diff filter by kind and namespace on new resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--diff-filter", `{"newResource": {"kinds": ["Service"], "namespaces": ["kapp-test"]}}`, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(serviceResourceYaml + configMapResourceyYaml)})

		expectedChange := []map[string]string{{
			"kind":            "Service",
			"name":            "redis-primary",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("diff filter on update operation", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--diff-filter", `{"ops": ["update"]}`, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(serviceResourceYaml + modifiedConfigMapResourceyYaml)})

		expectedChange := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      0 create, 0 delete, 1 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("diff filter delete resource with label", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name,
			"--diff-filter", `{"existingResource": {"labels": ["x=z"]}}`, "--json"},
			RunOpts{})

		expectedChange := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      0 create, 1 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 0 reconcile, 1 delete, 0 noop", out)
	})

	logger.Section("filter with or condition on new, existing resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--diff-filter", `{"or": [{"newResource": {"labels": ["x=z"]}},{"existingResource": {"labels": ["x!=z"]}}]}`, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(serviceResourceYaml + configMapResourceyYaml)})

		expectedChange := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("filter with and condition on new, existing resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--diff-filter", `{"and": [{"newResource": {"labels": ["x=y"]}},{"existingResource": {"labels": ["x=y"]}}]}`, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(modifiedServiceResourceYaml + configMapResourceyYaml)})

		expectedChange := []map[string]string{{
			"kind":            "Service",
			"name":            "redis-primary",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      0 create, 0 delete, 1 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("filter existing resource on labels", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--diff-filter", `{"not": {"existingResource": {"labels": ["x=y"]}}}`, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(serviceResourceYaml + modifiedConfigMapResourceyYaml)})

		expectedChange := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      0 create, 0 delete, 1 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("diff filter delete resource with name", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name,
			"--diff-filter", `{"existingResource": {"names": ["redis-primary", "redis-config"]}}`, "--json"},
			RunOpts{})

		expectedChange := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "Endpoints",
			"name":            "redis-primary",
			"namespace":       "kapp-test",
			"op":              "",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "Service",
			"name":            "redis-primary",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		validateChanges(t, resp.Tables, expectedChange, "Op:      0 create, 2 delete, 0 update, 1 noop, 0 exists",
			"Wait to: 0 reconcile, 3 delete, 0 noop", out)
	})
}
