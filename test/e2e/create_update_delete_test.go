// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestCreateUpdateDelete(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

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

	name := "test-create-update-delete"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
	})

	logger.Section("deploy update with 1 delete, 1 update, 1 create", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)

		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))

		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)
	})

	logger.Section("delete application", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config2", env.Namespace, kubectl)
	})
}
