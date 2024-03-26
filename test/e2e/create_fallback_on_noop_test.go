// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFallbackOnNoop(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	objYaml := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: precious-resource
  annotations:
    kapp.k14s.io/create-strategy: "fallback-on-update-or-noop"
data: {"importantFact":"true","releasedAt":null}
`
	rebaseRule := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
rebaseRules:
- ytt:
    overlayContractV1:
      overlay.yml: |
        #@ load("@ytt:overlay", "overlay")

        #@overlay/match by=overlay.all
        ---
        metadata:
          #@overlay/match missing_ok=True
          annotations:
            #@overlay/match missing_ok=True
            kapp.k14s.io/noop: ""
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: ConfigMap}
`
	name := "test-create-fallback-on-noop"
	name2 := "test-create-fallback-on-noop2"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", name2})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial app and assert it has its precious", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-y"},
			RunOpts{StdinReader: strings.NewReader(objYaml)})

		out := kapp.Run([]string{"inspect", "-a", name})
		require.Contains(t, out, "precious-resource")
	})

	logger.Section("deploy a second app with noop-strategy annotation", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2,
			"--existing-non-labeled-resources-check=false", "--dangerous-override-ownership-of-existing-resources", "-y", "--json"},
			RunOpts{StdinReader: strings.NewReader(objYaml + rebaseRule)})
	})

	logger.Section("assert the configmap still belongs to the first app, not the second", func() {
		out := kapp.Run([]string{"inspect", "-a", name})
		assert.Contains(t, out, "precious-resource")
		out = kapp.Run([]string{"inspect", "-a", name2})
		assert.NotContains(t, out, "precious-resource")
	})

	logger.Section("redeploy the second app without the rebase rule and it steals ownership of the configmap", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2,
			"--existing-non-labeled-resources-check=false", "--dangerous-override-ownership-of-existing-resources", "-y", "--json"},
			RunOpts{StdinReader: strings.NewReader(objYaml)})

		out := kapp.Run([]string{"inspect", "-a", name2})
		assert.Contains(t, out, "precious-resource")
		out = kapp.Run([]string{"inspect", "-a", name})
		assert.NotContains(t, out, "precious-resource")
	})
}
