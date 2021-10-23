// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/ghodss/yaml"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	config := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
rebaseRules:
- path: [data, delete]
  type: remove
  resourceMatchers:
  - kindNamespaceNameMatcher:
      kind: ConfigMap
      namespace: kapp-test
      name: first
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kapp-config
  labels:
    kapp.k14s.io/config: ""
data:
  config.yml: |
    apiVersion: kapp.k14s.io/v1alpha1
    kind: Config
    rebaseRules:
    - path: [data, keep]
      type: copy
      sources: [existing, new]
      resourceMatchers:
      - kindNamespaceNameMatcher:
          kind: ConfigMap
          namespace: kapp-test
          name: second
`

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: first
data:
  keep: ""
  delete: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
data:
  keep: ""
  delete: ""
` + config

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: first
data:
  keep: ""
  keep2: ""
  delete: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
data:
  keep: "replaced-value"
  keep2: ""
  delete: ""
` + config

	name := "test-config"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("initial deploy", func() {
		// Rebase rules are _only_ applied on the second run
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		firstData := NewPresentClusterResource("configmap", "first", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		require.Exactlyf(t, map[string]interface{}{"keep": "", "delete": ""}, firstData, "Expected value to be correct")

		secondData := NewPresentClusterResource("configmap", "second", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		require.Exactlyf(t, map[string]interface{}{"keep": "", "delete": ""}, secondData, "Expected value to be correct")
	})

	logger.Section("check rebases", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		firstData := NewPresentClusterResource("configmap", "first", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		require.Exactlyf(t, map[string]interface{}{"keep": "", "keep2": ""}, firstData, "Expected value to be correct")

		secondData := NewPresentClusterResource("configmap", "second", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		require.Exactlyf(t, map[string]interface{}{"keep": "", "keep2": "", "delete": ""}, secondData, "Expected value to be correct")
	})
}

func TestYttRebaseRule_ServiceAccountRebaseTokenSecret(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	// ServiceAccount controller appends secret named '${metadata.name}-token-${rand}'
	yaml1 := `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa-with-secrets
secrets:
- name: some-secret
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa-without-secrets`

	yaml2 := `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa-with-secrets
secrets:
- name: some-secret
- name: new-some-secret
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa-without-secrets`

	yaml3 := `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa-with-secrets
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa-without-secrets
secrets:
- name: some-secret`

	name := "test-config-ytt-rebase-sa-rebase"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	var generatedSecretName string

	logger.Section("initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 2, "Expected one set and one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0: %#v", secrets[0])

		generatedSecretName = secrets[1].(map[string]interface{})["name"].(string)
		require.True(t, strings.HasPrefix(generatedSecretName, "test-sa-with-secrets-token-"), "Expected generated secret at idx1: %#v", secrets[1])

		secrets = NewPresentClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 1, "Expected one set and one generated secret")
		require.True(t, strings.HasPrefix(secrets[0].(map[string]interface{})["name"].(string), "test-sa-without-secrets-token-"), "Expected generated secret at idx0: %#v", secrets[0])
	})

	ensureDeploysWithNoChanges := func(yamlContent string) {
		for i := 0; i < 3; i++ { // Try doing it a few times
			logger.Section("deploy with no changes as rebase rule should retain generated secrets", func() {
				out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json", "-c"},
					RunOpts{IntoNs: true, StdinReader: strings.NewReader(yamlContent)})

				resp := uitest.JSONUIFromBytes(t, []byte(out))
				expected := []map[string]string{}

				require.Exactlyf(t, expected, resp.Tables[0].Rows, "Expected to see correct changes, but did not")
				require.Equalf(t, "Op:      0 create, 0 delete, 0 update, 0 noop", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
			})
		}
	}

	ensureDeploysWithNoChanges(yaml1)

	logger.Section("deploy with additional secret, but retain existing generated secret", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 3, "Expected one set and one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0")
		require.Exactlyf(t, map[string]interface{}{"name": "new-some-secret"}, secrets[1], "Expected provided secret at idx1")
		require.Exactlyf(t, map[string]interface{}{"name": generatedSecretName}, secrets[2], "Expected previous generated secret at idx2")
	})

	ensureDeploysWithNoChanges(yaml2)

	logger.Section("deploy with flipped secrets", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 1, "Expected one set and one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": generatedSecretName}, secrets[0], "Expected previous generated secret at idx0")

		secrets = NewPresentClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 2, "Expected one set and one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0")
		require.True(t, strings.HasPrefix(secrets[1].(map[string]interface{})["name"].(string), "test-sa-without-secrets-token-"), "Expected generated secret at idx1: %#v", secrets[1])
	})

	ensureDeploysWithNoChanges(yaml3)
}

func TestYttRebaseRule_OverlayContractV1(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	config := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
- ytt:
    overlayContractV1:
      overlay.yml: |
        #@ load("@ytt:data", "data")
        #@ load("@ytt:overlay", "overlay")

        #@overlay/match by=overlay.all
        ---
        data:
          #! will be visible in data.values._current in next ytt rebase
          #@overlay/match missing_ok=True
          changed_in_rebase_rule: "1"
  resourceMatchers:
  - allMatcher: {}

- ytt:
    overlayContractV1:
      overlay.yml: |
        #@ load("@ytt:data", "data")
        #@ load("@ytt:yaml", "yaml")
        #@ load("@ytt:overlay", "overlay")

        #@overlay/match by=overlay.all
        ---
        data:
          #! expected to find this key from prev rebase rule
          changed_in_rebase_rule: "2"

          #@ if not hasattr(data.values.existing.data, "values"):

          #! this would run on the first rebase
          #@overlay/match missing_ok=True
          values: #@ yaml.encode(data.values)

          #@ else:

          #! this would run on the second rebase since existing
          #! resource contains prev applied values
          #@overlay/match missing_ok=True
          values: #@ data.values.existing.data.values

          #@ end

  resourceMatchers:
  - allMatcher: {}
`

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
data:
  key1: val1`

	name := "test-config-ytt-rebase-overlay-contract-v1"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	var cm ClusterResource

	logger.Section("initial deploy (rebase does not run since there is no existing resource)", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(config + yaml1)})

		cm = NewPresentClusterResource("configmap", "test-cm", env.Namespace, kubectl)
		data := cm.RawPath(ctlres.NewPathFromStrings([]string{"data"})).(map[string]interface{})

		require.Equal(t, map[string]interface{}{"key1": "val1"}, data)
	})

	var expectedDataStr string

	logger.Section("second deploy (rebase runs)", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(config + yaml1)})

		expectedDataStr = asYAML(t, map[string]interface{}{
			"key1": "val1",
			// Following fields are accessible via data.values inside ytt:
			// - data.values.existing: resource from live cluster
			// - data.values.new: resource from config (post-prep)
			// - data.values._current: resource after previous rebase rules already applied
			"values": asYAML(t, map[string]interface{}{
				"existing": func() interface{} {
					raw := cm.Raw()
					metadata := raw["metadata"].(map[string]interface{})
					anns := metadata["annotations"].(map[string]interface{})
					delete(anns, "kapp.k14s.io/identity")
					return raw
				}(),
				"_current": func() interface{} {
					raw := cm.Raw()
					metadata := raw["metadata"].(map[string]interface{})
					delete(metadata, "annotations")
					data := raw["data"].(map[string]interface{})
					data["changed_in_rebase_rule"] = "1"
					return raw
				}(),
				"new": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": "test-cm",
						// Namespace is added as part of kapp preparation step for input resources
						"namespace": env.Namespace,
						// These labels are added as part of kapp preparation step for input resources
						"labels": map[string]interface{}{
							"kapp.k14s.io/app":         cm.Labels()["kapp.k14s.io/app"],
							"kapp.k14s.io/association": cm.Labels()["kapp.k14s.io/association"],
						},
					},
					"data": map[string]interface{}{
						"key1": "val1",
					},
				},
			}),
			"changed_in_rebase_rule": "2",
		})

		cm = NewPresentClusterResource("configmap", "test-cm", env.Namespace, kubectl)
		data := cm.RawPath(ctlres.NewPathFromStrings([]string{"data"})).(map[string]interface{})

		require.Equal(t, expectedDataStr, asYAML(t, data))
	})

	logger.Section("third deploy with no changes (rebase runs)", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(config + yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		expected := []map[string]string{}

		require.Exactlyf(t, expected, resp.Tables[0].Rows, "Expected to see correct changes, but did not")
		require.Equalf(t, "Op:      0 create, 0 delete, 0 update, 0 noop", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")

		cm = NewPresentClusterResource("configmap", "test-cm", env.Namespace, kubectl)
		data := cm.RawPath(ctlres.NewPathFromStrings([]string{"data"})).(map[string]interface{})

		require.Equal(t, expectedDataStr, asYAML(t, data))
	})
}

func asYAML(t *testing.T, val interface{}) string {
	bs, err := yaml.Marshal(val)
	require.NoError(t, err)
	return string(bs)
}
