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

func TestAppSuffix_AppExistsWithoutSuffix(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-app-suffix-app-exists"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	existingApp := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-app-suffix-app-exists
  labels:
    kapp.k14s.io/is-app: ""
data:
  spec: '{"labelKey":"kapp.k14s.io/app","labelValue":"1641592268838201000","lastChange":{"startedAt":"0001-01-01T00:00:00Z","finishedAt":"0001-01-01T00:00:00Z"}}'
`

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
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
`

	yaml2 := `
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

	logger.Section("deploy with 1 delete, 1 update, 1 create", func() {
		kubectl.RunWithOpts([]string{"apply", "-f", "-"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(existingApp)})
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)

		cleanUp()
	})

	logger.Section("rename", func() {
		newName := "test-app-suffix-app-exists-new"

		kubectl.RunWithOpts([]string{"apply", "-f", "-"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(existingApp)})
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		kapp.Run([]string{"rename", "-a", name, "--new-name", newName})
		NewPresentClusterResource("configmap", newName+app.AppSuffix, env.Namespace, kubectl)

		kapp.Run([]string{"delete", "-a", newName})
	})

	logger.Section("delete", func() {
		kubectl.RunWithOpts([]string{"apply", "-f", "-"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(existingApp)})
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		kapp.Run([]string{"delete", "-a", name})

		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
	})

	logger.Section("configmap with suffix exists and not marked as a kapp-app", func() {
		fqName := name + app.AppSuffix

		NewClusterResource(t, "configmap", fqName, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", fqName, env.Namespace, kubectl)

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.Containsf(t, err.Error(), "kapp: Error:", "did not contain parseable app metadata")

		RemoveClusterResource(t, "configmap", fqName, env.Namespace, kubectl)
	})

	logger.Section("does not migrate if USE_EXISTING_CONFIGMAP_NAME=True", func() {
		os.Setenv("USE_EXISTING_CONFIGMAP_NAME", "True")

		kubectl.RunWithOpts([]string{"apply", "-f", "-"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(existingApp)})
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		os.Unsetenv("USE_EXISTING_CONFIGMAP_NAME")
		cleanUp()
	})
}
