// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/k14s/kapp/pkg/kapp/app"
)

var (
	yaml1 string = `
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

	yaml2 string = `
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
)

func TestAppSuffix_CreateUpdateDelete(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-create-update-delete"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
	})

	logger.Section("deploy update with 1 delete, 1 update, 1 create", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
	})

	logger.Section("delete application", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})

		NewMissingClusterResource(t, "configmap", name+app.AppSuffix, env.Namespace, kubectl)
	})
}

func TestAppSuffix_Rename(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-app-suffix-rename"
	newName := "test-app-suffix-rename-new"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", newName})
	}

	cleanUp()
	defer cleanUp()

	kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)

	kapp.Run([]string{"rename", "-a", name, "--new-name", newName})
	NewPresentClusterResource("configmap", newName+app.AppSuffix, env.Namespace, kubectl)
}

func TestAppSuffix_AppExistsWithoutSuffix(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kappOld := Kapp{t, env.Namespace, env.KappBinaryPathOld, logger}
	kappNew := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-app-suffix-app-exists"
	cleanUp := func() {
		kappNew.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("configmap marked as kapp-app", func() {
		kappOld.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name, env.Namespace, kubectl)

		kappNew.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)

		kappNew.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		NewPresentClusterResource("configmap", name+app.AppSuffix, env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)

		cleanUp()
	})
}
