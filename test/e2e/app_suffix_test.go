// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"

	"carvel.dev/kapp/pkg/kapp/app"
	"github.com/stretchr/testify/require"
)

var yaml1 = `
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
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
`

var yaml2 = `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
`

func TestAppSuffix_AppExists_MigrationEnabled(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-app-suffix-app-exists"
	newName := "test-app-suffix-app-exists-new"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy with 1 delete, 1 update, 1 create", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		// update and migrate
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		c := NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

		cleanUp()
	})

	logger.Section("inspect", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		kapp.Run([]string{"inspect", "-a", name})
	})

	logger.Section("rename", func() {
		kapp.Run([]string{"rename", "-a", name, "--new-name", newName})
		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", newName+app.AppSuffix, env.Namespace, kubectl)
	})

	logger.Section("delete", func() {
		kapp.Run([]string{"delete", "-a", newName})

		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", newName, env.Namespace, kubectl)
	})

	logger.Section("name contains app suffix for migrated app", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name + app.AppSuffix}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		// double migration shouldn't have happened
		NewMissingClusterResource(t, "configmap", name+app.AppSuffix+app.AppSuffix, env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		kapp.Run([]string{"delete", "-a", name})
	})

	// Migrated apps are supported even when migration is disabled
	logger.Section("migration disabled with already migrated app", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)

		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "False")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)

	})

	os.Unsetenv("KAPP_FQ_CONFIGMAP_NAMES")
}
