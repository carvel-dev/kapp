// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"path"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

func TestAppGroupCreateUpdateDelete(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	app1 := `
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

	app2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
`

	appTwoUpdated := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value2
`
	appGroupDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}

	appOneDir := "app-1"
	appTwoDir := "app-2"
	err = os.Mkdir(path.Join(appGroupDir, appOneDir), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(path.Join(appGroupDir, appTwoDir), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path.Join(appGroupDir, appOneDir, "config.yml"), []byte(app1), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path.Join(appGroupDir, appTwoDir, "config.yml"), []byte(app2), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	name := "test-create-update-delete-app-group"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()
	defer os.RemoveAll(appGroupDir)

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"app-group", "deploy", "-g", name, "--directory", appGroupDir}, RunOpts{IntoNs: true})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)

		listedApps := kapp.Run([]string{"ls", "--json"})

		expectedAppsList := []map[string]string{
			{
				"last_change_age":        "<replaced>",
				"last_change_successful": "true",
				"name":                   name + "-" + appOneDir,
				"namespaces":             env.Namespace,
			},
			{
				"last_change_age":        "<replaced>",
				"last_change_successful": "true",
				"name":                   name + "-" + appTwoDir,
				"namespaces":             env.Namespace,
			},
		}

		resp := uitest.JSONUIFromBytes(t, []byte(listedApps))

		require.Equalf(t, expectedAppsList, replaceLastChangeAge(resp.Tables[0].Rows), "Expected to match")
	})

	logger.Section("deploy app-group with an update", func() {
		err = os.WriteFile(path.Join(appGroupDir, appTwoDir, "config.yml"), []byte(appTwoUpdated), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		kapp.RunWithOpts([]string{"app-group", "deploy", "-g", name, "--directory", appGroupDir}, RunOpts{IntoNs: true})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))

		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		listedApps := kapp.Run([]string{"ls", "--json"})

		expectedAppsList := []map[string]string{
			{
				"last_change_age":        "<replaced>",
				"last_change_successful": "true",
				"name":                   name + "-" + appOneDir,
				"namespaces":             env.Namespace,
			},
			{
				"last_change_age":        "<replaced>",
				"last_change_successful": "true",
				"name":                   name + "-" + appTwoDir,
				"namespaces":             env.Namespace,
			},
		}

		resp := uitest.JSONUIFromBytes(t, []byte(listedApps))

		require.Equalf(t, expectedAppsList, replaceLastChangeAge(resp.Tables[0].Rows), "Expected to match")
	})

	logger.Section("delete app-group", func() {
		kapp.RunWithOpts([]string{"app-group", "delete", "-g", name}, RunOpts{})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config", env.Namespace, kubectl)

		listedApps := kapp.Run([]string{"ls", "--json"})

		expectedAppsList := []map[string]string{}

		resp := uitest.JSONUIFromBytes(t, []byte(listedApps))

		require.Equalf(t, expectedAppsList, replaceLastChangeAge(resp.Tables[0].Rows), "Expected to match")
	})
}
