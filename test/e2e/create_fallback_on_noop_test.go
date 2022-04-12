// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFallbackOnNoop(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	objNs := env.Namespace + "-create-fallback-on-noop"
	nsYaml := strings.Replace(`
---
apiVersion: v1
kind: Namespace
metadata:
  name: __ns__
`, "__ns__", objNs, -1)
	objYaml := strings.Replace(`
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: precious-resource
  namespace: __ns__
  annotations:
    kapp.k14s.io/create-strategy: "fallback-on-update-or-noop"
data: {"importantFact":"true","releasedAt":null}
`, "__ns__", objNs, -1)

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
	name2 := "test-2"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", name2})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial ns and package", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-y"},
			RunOpts{AllowError: false, StdinReader: strings.NewReader(nsYaml + objYaml)})
		assert.NoError(t, err)
	})

	logger.Section("assert theres a configmap that belongs to the app", func() {
		out := kapp.Run([]string{"inspect", "-a", name})
		assert.Contains(t, out, "precious-resource")
	})

	logger.Section("deploy a second app with noop-strategy annotation", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2,
			"--existing-non-labeled-resources-check=false", "--dangerous-override-ownership-of-existing-resources", "-y", "--json"},
			RunOpts{AllowError: false, StdinReader: strings.NewReader(objYaml + rebaseRule)})
		assert.NoError(t, err)
	})

	logger.Section("assert the configmap still belongs to the first app, not the second", func() {
		out := kapp.Run([]string{"inspect", "-a", name})
		assert.Contains(t, out, "precious-resource")
		out = kapp.Run([]string{"inspect", "-a", name2})
		assert.NotContains(t, out, "precious-resource")
	})

	logger.Section("redeploy the second app without the rebase rule and it steals ownership of the configmap", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2,
			"--existing-non-labeled-resources-check=false", "--dangerous-override-ownership-of-existing-resources", "-y", "--json"},
			RunOpts{AllowError: false, StdinReader: strings.NewReader(objYaml)})
		assert.NoError(t, err)

		out := kapp.Run([]string{"inspect", "-a", name2})
		assert.Contains(t, out, "precious-resource")
		out = kapp.Run([]string{"inspect", "-a", name})
		assert.NotContains(t, out, "precious-resource")
	})
}
