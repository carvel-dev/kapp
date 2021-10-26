// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestTransientResourceInspectDelete(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-svc
spec:
  selector:
    app: redis-svc
  ports:
  - name: http
    port: 80
`

	name := "test-transient-resource-inspect-delete"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("inspect shows transient resource", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "--json"}, RunOpts{})
		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if hasEndpointSlice(respRows) {
			respRows = removeEndpointSliceNameSuffix(respRows)
			expected = append(expected, map[string]string{
				"age":             "<replaced>",
				"conditions":      "",
				"kind":            "EndpointSlice",
				"name":            "redis-svc",
				"namespace":       "kapp-test",
				"owner":           "cluster",
				"reconcile_info":  "",
				"reconcile_state": "ok",
			})
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})

	logger.Section("delete includes transient resource", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})
		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if hasEndpointSlice(respRows) {
			respRows = removeEndpointSliceNameSuffix(respRows)
			expected = append(expected, map[string]string{
				"age":             "<replaced>",
				"op":              "",
				"op_strategy":     "",
				"wait_to":         "delete",
				"conditions":      "",
				"kind":            "EndpointSlice",
				"name":            "redis-svc",
				"namespace":       "kapp-test",
				"reconcile_info":  "",
				"reconcile_state": "ok",
			})
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})
}

func TestTransientResourceSwitchToNonTransient(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-svc
spec:
  selector:
    app: redis-svc
  ports:
  - name: http
    port: 80
`

	yaml2 := yaml1 + `
---
apiVersion: v1
kind: Endpoints
metadata:
  name: redis-svc
`

	name := "test-transient-resource-switch"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy to change transient to non-transient", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "update",
			"op_strategy":     "",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})

	logger.Section("delete with previously transient resource (now non-transient)", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})
		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if hasEndpointSlice(respRows) {
			respRows = removeEndpointSliceNameSuffix(respRows)
			expected = append(expected, map[string]string{
				"age":             "<replaced>",
				"op":              "",
				"op_strategy":     "",
				"wait_to":         "delete",
				"conditions":      "",
				"kind":            "EndpointSlice",
				"name":            "redis-svc",
				"namespace":       "kapp-test",
				"reconcile_info":  "",
				"reconcile_state": "ok",
			})
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})
}

func removeEndpointSliceNameSuffix(result []map[string]string) []map[string]string {
	for i, row := range result {
		if row["kind"] == "EndpointSlice" && len(row["name"]) > 0 {
			lastIndexOfDash := strings.LastIndex(row["name"], "-")
			row["name"] = row["name"][:lastIndexOfDash]
		}
		result[i] = row
	}
	return result
}

func hasEndpointSlice(result []map[string]string) bool {
	for _, row := range result {
		if row["kind"] == "EndpointSlice" {
			return true
		}
	}
	return false
}
