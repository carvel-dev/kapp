// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateFallbackOnReplace(t *testing.T) {
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
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
  annotations:
    kapp.k14s.io/update-strategy: fallback-on-replace
spec:
  clusterIP: None
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
`

	name := "test-update-fallback-on-replace"
	objKind := "service"
	objName := "redis-primary"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy update to service that changes immutable field spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		require.NotEqual(t, prev.UID(), curr.UID(), "Expected object to be replaced, but found same UID")
	})

	logger.Section("deploy update to service that does not set spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		require.Equal(t, prev.UID(), curr.UID(), "Expected object to be rebased, but found different UID")
	})
}

func TestUpdateAlwaysReplace(t *testing.T) {
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
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
  annotations:
    kapp.k14s.io/update-strategy: always-replace
spec:
  clusterIP: None
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
`

	name := "test-update-always-replace"
	objKind := "service"
	objName := "redis-primary"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy update to service that changes immutable field spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		require.NotEqual(t, prev.UID(), curr.UID(), "Expected object to be replaced, but found same UID")
	})

	logger.Section("deploy update to service that does not set spec.clusterIP", func() {
		prev := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		curr := NewPresentClusterResource(objKind, objName, env.Namespace, kubectl)

		require.Equal(t, prev.UID(), curr.UID(), "Expected object to be rebased, but found different UID")
	})
}
