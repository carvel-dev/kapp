// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestDeleteOrphan(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  annotations:
    kapp.k14s.io/delete-strategy: orphan
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
  key: value
`

	name := "test-delete-orphan"
	nameAnother := "test-delete-orphan-another"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", nameAnother})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
	})

	logger.Section("deploy without resource to orphan", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--dangerous-allow-empty-list-of-resources"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader("")})
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
	})

	logger.Section("deploy another app to adopt", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", nameAnother},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
	})

	logger.Section("delete original app", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
	})

	logger.Section("undo orphaning behaviour and delete another app", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", nameAnother},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"delete", "-a", nameAnother}, RunOpts{})
		NewMissingClusterResource(t, "configmap", "redis-config", env.Namespace, kubectl)
	})
}
