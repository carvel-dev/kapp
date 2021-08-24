// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestLabelFilter(t *testing.T) {
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
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  labels:
    x: "z"
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
`

	name := "test-label-filter"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--filter-labels", "x=y,x=z"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config2", env.Namespace, kubectl)
	})

	logger.Section("delete application", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name, "--filter-labels", "x=y"}, RunOpts{})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config2", env.Namespace, kubectl)
	})

	logger.Section("delete application", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name, "--filter-labels", "x=z"}, RunOpts{})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config2", env.Namespace, kubectl)
	})

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--filter-labels", "x!=y"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)
	})

}
