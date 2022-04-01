// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestDeleteInoperable(t *testing.T) {

	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
--- 
apiVersion: v1
kind: Namespace
metadata: 
  name: default

`

	name := "test-delete-inoperable"
	nameAnother := "test-delete-inoperable-another"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", nameAnother})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})

	logger.Section("delete initial", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})

	logger.Section("deploy again with same app name", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})

	logger.Section("deploy without resource to orphan", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--dangerous-allow-empty-list-of-resources"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader("")})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})

	logger.Section("deploy another app to adopt", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", nameAnother},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})

	logger.Section("delete original app", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})

	logger.Section("delete another app", func() {
		kapp.RunWithOpts([]string{"delete", "-a", nameAnother}, RunOpts{})
		NewPresentClusterResource("namespace", "default", env.Namespace, kubectl)
	})
}
