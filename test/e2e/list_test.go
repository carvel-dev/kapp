// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
	"time"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestAppListing(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
  labels:
    x: "y"
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
`
	yaml2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  labels:
    x: "z"
data:
  key: value
`

	name := "test-app-1"
	name2 := "test-app-2"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", name2})
	}

	cleanUp()
	defer cleanUp()
	logger.Section("App listing and filter label", func() {
		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
			"--labels", "x=y"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2,
			"--labels", "a=b"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		listedAppsWithAge, _ := kapp.RunWithOpts([]string{"ls", "--filter-age", "2m+", "--json"}, RunOpts{Interactive: true})

		response := uitest.JSONUIFromBytes(t, []byte(listedAppsWithAge))

		if len(response.Tables) > 0 {
			require.Empty(t, response.Tables[0].Rows, "Expected table rows to empty")
		}

		time.Sleep(2 * time.Second)

		listedApps, _ := kapp.RunWithOpts([]string{"ls", "--filter-age", "2s+", "--json"}, RunOpts{Interactive: true})

		expectedAppsList := []map[string]string{{
			"last_change_successful": "true",
			"name":                   "test-app-1",
			"namespaces":             "kapp-test",
		}, {
			"last_change_successful": "true",
			"name":                   "test-app-2",
			"namespaces":             "kapp-test",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(listedApps))

		validateAppListChanges(t, resp.Tables, expectedAppsList, listedApps)

		filteredApps, _ := kapp.RunWithOpts([]string{"ls", "--filter-labels", "x=y", "--json"}, RunOpts{Interactive: true})

		expectedFilteredApps := []map[string]string{{
			"last_change_successful": "true",
			"name":                   "test-app-1",
			"namespaces":             "kapp-test",
		}}

		resp2 := uitest.JSONUIFromBytes(t, []byte(filteredApps))

		validateAppListChanges(t, resp2.Tables, expectedFilteredApps, filteredApps)
	})
}
