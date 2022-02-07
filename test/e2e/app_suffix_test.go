// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/k14s/kapp/pkg/kapp/app"
	"github.com/stretchr/testify/require"
)

var yaml1 string = `
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

var yaml2 string = `
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

		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "True")
		// update and migrate
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		c := NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.Labels(), app.KappIsConfigmapMigratedLabelKey)

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

		cleanUp()
		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", newName, env.Namespace, kubectl)
	})

	logger.Section("name already contains app suffix", func() {
		existingName := name + app.AppSuffix

		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "False")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", existingName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", existingName, env.Namespace, kubectl)

		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", existingName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewMissingClusterResource(t, "configmap", existingName, env.Namespace, kubectl)
		c := NewPresentClusterResource("configmap", existingName+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.Labels(), app.KappIsConfigmapMigratedLabelKey)

		kapp.Run([]string{"delete", "-a", existingName})
	})

	// if a user has disabled the migration on an app that was already migrated
	// the error we expect is that the resources belong to a different app
	// - there will be a documented way to recover from this if desired
	logger.Section("migration disabled with already migrated app", func() {
		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "False")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "False")
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true,
			AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.Containsf(t, err.Error(), "kapp: Error: Ownership errors:", "Expected resources to be associated with a different app")
		// the old name configmap is created
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		RemoveClusterResource(t, "configmap", name, env.Namespace, kubectl)

		os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "True")
		cleanUp()
	})

	os.Unsetenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES")
}

func TestAppSuffix_ConfigmapExists_MigrationEnabled(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-app-suffix-configmap-exists"
	fqName := name + app.AppSuffix

	cleanUp := func() {
		RemoveClusterResource(t, "configmap", name, env.Namespace, kubectl)
		RemoveClusterResource(t, "configmap", fqName, env.Namespace, kubectl)
	}

	cleanUp()
	defer cleanUp()

	os.Setenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES", "True")

	logger.Section("without suffix and not marked as a kapp-app", func() {
		NewClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.NoError(t, err)
		cleanUp()
	})

	logger.Section("with suffix and not marked as a kapp-app", func() {
		NewClusterResource(t, "configmap", fqName, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", fqName, env.Namespace, kubectl)

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "did not contain parseable app metadata", "Expected app to not be parsable")
	})

	os.Unsetenv("KAPP_EXPERIMENTAL_FQ_CONFIGMAP_NAMES")
}
