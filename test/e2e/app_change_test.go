// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestAppChange(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml := `
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
    tier: %s
`

	name := "test-app-change"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy app", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, "backend"))})
	})

	logger.Section("deploy with changes", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, "frontend"))})
	})

	logger.Section("app change list", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 2, len(resp.Tables[0].Rows), "Expected to have 2 app-changes")
		require.Equal(t, "update: Op: 0 create, 0 delete, 1 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[0]["description"], "Expected description to match")
		require.Equal(t, "update: Op: 1 create, 0 delete, 0 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[1]["description"], "Expected description to match")
	})

	logger.Section("app change list filter with before flag", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--before", time.Now().Add(1 * time.Second).Format(time.RFC3339), "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 2, len(resp.Tables[0].Rows), "Expected to have 2 app-changes")
		require.Equal(t, "update: Op: 0 create, 0 delete, 1 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[0]["description"], "Expected description to match")
		require.Equal(t, "update: Op: 1 create, 0 delete, 0 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[1]["description"], "Expected description to match")

		out2, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--before", time.Now().Add(-1 * time.Minute).Format(time.RFC3339), "--json"}, RunOpts{})

		resp2 := uitest.JSONUIFromBytes(t, []byte(out2))

		require.Equal(t, 0, len(resp2.Tables[0].Rows), "Expected to have 0 app-changes")
	})

	logger.Section("app change list filter with after flag", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--after", time.Now().Add(24 * time.Hour).Format("2006-01-02"), "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 0, len(resp.Tables[0].Rows), "Expected to have 0 app-changes")

		out2, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--after", time.Now().Add(-1 * time.Minute).Format("2006-01-02"), "--json"}, RunOpts{})

		resp2 := uitest.JSONUIFromBytes(t, []byte(out2))

		require.Equal(t, 2, len(resp2.Tables[0].Rows), "Expected to have 2 app-changes")
	})
}

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

type usedGK struct {
	Group string `yaml:"Group"`
	Kind  string `yaml:"Kind"`
}

type lastChange struct {
	Namespaces []string `yaml:"namespaces"`
}

type yamlSubset struct {
	LastChange lastChange `yaml:"lastChange"`
	UsedGKs    []usedGK   `yaml:"usedGKs"`
}
