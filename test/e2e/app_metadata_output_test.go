// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestAppKindChangeWithMetadataOutput(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  redis-config: |
    maxmemory 3mb
    maxmemory-policy allkeys-lru
`

	yaml2 := `
---
apiVersion: v1
kind: Secret
metadata:
  name: kapp-secret-1
  namespace: kapp-namespace-2
`

	name := "test-app-change"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	firstDeploy, err := os.CreateTemp(os.TempDir(), "output1")
	assert.NoError(t, err)
	secondDeploy, err := os.CreateTemp(os.TempDir(), "output2")
	assert.NoError(t, err)
	thirdDeploy, err := os.CreateTemp(os.TempDir(), "output3")
	assert.NoError(t, err)

	defer func() {
		os.Remove(firstDeploy.Name())
		os.Remove(secondDeploy.Name())
	}()

	logger.Section("deploy app", func() {
		kapp.RunWithOpts([]string{"deploy", "--app-metadata-file-output", firstDeploy.Name(), "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy with changes", func() {
		kapp.RunWithOpts([]string{"deploy", "--app-metadata-file-output", secondDeploy.Name(), "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
	})

	logger.Section("deploy with no changes", func() {
		kapp.RunWithOpts([]string{"deploy", "--app-metadata-file-output", thirdDeploy.Name(), "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
	})

	configMapFirstDeploy, err := os.ReadFile(firstDeploy.Name())
	assert.NoError(t, err)

	firstConfigMap := yamlSubset{}
	require.NoError(t, yaml.Unmarshal(configMapFirstDeploy, &firstConfigMap))
	require.Equal(t, yamlSubset{LastChange: lastChange{Namespaces: []string{env.Namespace}}, UsedGKs: []usedGK{{Group: "", Kind: "ConfigMap"}}}, firstConfigMap)

	configMapSecondDeploy, err := os.ReadFile(secondDeploy.Name())
	assert.NoError(t, err)

	secondConfigMap := yamlSubset{}
	require.NoError(t, yaml.Unmarshal(configMapSecondDeploy, &secondConfigMap))
	require.Equal(t, yamlSubset{LastChange: lastChange{Namespaces: []string{env.Namespace}}, UsedGKs: []usedGK{{Group: "", Kind: "Secret"}}}, secondConfigMap)

	configMapThirdDeploy, err := os.ReadFile(thirdDeploy.Name())
	assert.NoError(t, err)

	thirdConfigMap := yamlSubset{}
	require.NoError(t, yaml.Unmarshal(configMapThirdDeploy, &thirdConfigMap))
	require.Equal(t, yamlSubset{LastChange: lastChange{Namespaces: []string{env.Namespace}}, UsedGKs: []usedGK{{Group: "", Kind: "Secret"}}}, thirdConfigMap)
}
