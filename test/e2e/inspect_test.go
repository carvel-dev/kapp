// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestInspect(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
`

	name := "test-inspect"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("plain inspect", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-primary",
			"namespace":       "kapp-test",
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-primary",
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
				"name":            "redis-primary",
				"namespace":       "kapp-test",
				"owner":           "cluster",
				"reconcile_info":  "",
				"reconcile_state": "ok",
			})
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})

	logger.Section("tree inspect", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "-t", "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-primary",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            " L redis-primary",
			"namespace":       "kapp-test",
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if hasEndpointSlice(respRows) {
			respRows = removeEndpointSliceNameSuffix(respRows)
			expected = append(expected, map[string]string{
				"age":             "<replaced>",
				"conditions":      "",
				"kind":            "EndpointSlice",
				"name":            " L redis-primary",
				"namespace":       "kapp-test",
				"owner":           "cluster",
				"reconcile_info":  "",
				"reconcile_state": "ok",
			})
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})
}
