// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
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
	minorVersion, err := getServerMinorVersion()
	require.NoErrorf(t, err, "Error getting k8s server minor version")

	if minorVersion >= 24 {
		t.Skip("Automatic creation of service account token is turned off in k8s v1.24.0+")
	}

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
		require.Len(t, secrets, 1, "Expected one generated secret")
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
				require.Equalf(t, "Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
			})
		}
	}

	ensureDeploysWithNoChanges(yaml1)

	logger.Section("deploy with additional secret, but retain existing generated secret", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 3, "Expected two set and one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0")
		require.Exactlyf(t, map[string]interface{}{"name": "new-some-secret"}, secrets[1], "Expected provided secret at idx1")
		require.Exactlyf(t, map[string]interface{}{"name": generatedSecretName}, secrets[2], "Expected previous generated secret at idx2")
	})

	ensureDeploysWithNoChanges(yaml2)

	logger.Section("deploy with flipped secrets", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 1, "Expected one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": generatedSecretName}, secrets[0], "Expected previous generated secret at idx0")

		secrets = NewPresentClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.Len(t, secrets, 2, "Expected one set and one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0")
		require.True(t, strings.HasPrefix(secrets[1].(map[string]interface{})["name"].(string), "test-sa-without-secrets-token-"), "Expected generated secret at idx1: %#v", secrets[1])
	})

	ensureDeploysWithNoChanges(yaml3)
}

func TestYttRebaseRule_ServiceAccountRebaseTokenSecret_OpenShift(t *testing.T) {
	minorVersion, err := getServerMinorVersion()
	require.NoErrorf(t, err, "Error getting k8s server minor version")

	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	// Openshift appends secret and imagePullSecret named '${metadata.name}-dockercfg-${rand}'
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

	logger.Section("initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		patchSAWithSecrets := `[{ "op": "add", "path": "/imagePullSecrets", "value": [{ "name": "test-sa-with-secrets-dockercfg-<rand>"}]},
			{ "op": "add", "path": "/secrets/-", "value": { "name": "test-sa-with-secrets-dockercfg-<rand>"}}]`

		patchSAWithoutSecrets := `[{ "op": "add", "path": "/imagePullSecrets", "value": [{ "name": "test-sa-without-secrets-dockercfg-<rand>"}]},
			{ "op": "add", "path": "/secrets/-", "value": { "name": "test-sa-without-secrets-dockercfg-<rand>"}}]`

		if minorVersion >= 24 {
			patchSAWithoutSecrets = `[{ "op": "add", "path": "/imagePullSecrets", "value": [{ "name": "test-sa-without-secrets-dockercfg-<rand>"}]},
			{ "op": "add", "path": "/secrets", "value": [{ "name": "test-sa-without-secrets-dockercfg-<rand>"}]}]`
		}

		// Mock Openshift behavior by adding additional secrets and image pull secrets
		PatchClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, strings.ReplaceAll(patchSAWithSecrets, "<rand>", RandomString(5)), kubectl)
		PatchClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, strings.ReplaceAll(patchSAWithoutSecrets, "<rand>", RandomString(5)), kubectl)

		serviceAccount := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl)

		secrets := serviceAccount.RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.GreaterOrEqual(t, len(secrets), 2, "Expected one set and at least one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0: %#v", secrets[0])

		secrets = NewPresentClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.GreaterOrEqual(t, len(secrets), 1, "Expected at least one generated secret")
	})

	ensureDeploysWithNoChanges := func(yamlContent string) {
		for i := 0; i < 3; i++ { // Try doing it a few times
			logger.Section("deploy with no changes as rebase rule should retain generated secrets", func() {
				out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--json"},
					RunOpts{IntoNs: true, StdinReader: strings.NewReader(yamlContent)})

				resp := uitest.JSONUIFromBytes(t, []byte(out))
				expected := []map[string]string{}

				require.Exactlyf(t, expected, resp.Tables[0].Rows, "Expected to see correct changes, but did not")
				require.Equalf(t, "Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
			})
		}
	}

	ensureDeploysWithNoChanges(yaml1)

	logger.Section("deploy with additional secret, but retain existing generated secret", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.GreaterOrEqual(t, len(secrets), 3, "Expected two set and at least one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0")
		require.Exactlyf(t, map[string]interface{}{"name": "new-some-secret"}, secrets[1], "Expected provided secret at idx1")

		secrets = NewPresentClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.GreaterOrEqual(t, len(secrets), 1, "Expected at least one generated secret")
	})

	ensureDeploysWithNoChanges(yaml2)

	logger.Section("deploy with flipped secrets", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa-with-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.GreaterOrEqual(t, len(secrets), 1, "Expected at least one generated secret")

		secrets = NewPresentClusterResource("serviceaccount", "test-sa-without-secrets", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		require.GreaterOrEqual(t, len(secrets), 2, "Expected one set and at least one generated secret")
		require.Exactlyf(t, map[string]interface{}{"name": "some-secret"}, secrets[0], "Expected provided secret at idx0")
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
		require.Equalf(t, "Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")

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

func TestDefaultConfig_ExcludeDiffAgainstExistingStatus(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml := `
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.kapp.example
spec:
  group: kapp.example
  names:
    kind: Test
    plural: tests
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        type: object
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	updatedYaml := `
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.kapp.example
spec:
  group: kapp.example
  names:
    kind: Test
    plural: tests
  scope: Namespaced
  versions:
  - name: v1beta1
    schema:
      openAPIV3Schema:
        type: object
    served: true
    storage: true
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        type: object
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`

	expectedChangesForUpdate := `
@@ update customresourcedefinition/tests.kapp.example (apiextensions.k8s.io/v1) cluster @@
  ...
-linesss-   versions:
-linesss-   - name: v1alpha1
-linesss-   - name: v1beta1
-linesss-     schema:
-linesss-       openAPIV3Schema:
  ...
-linesss-     storage: true
-linesss-   - name: v1alpha1
-linesss-     schema:
-linesss-       openAPIV3Schema:
-linesss-         type: object
-linesss- `

	expectedChangesForExternalUpdate := `
@@ update customresourcedefinition/tests.kapp.example (apiextensions.k8s.io/v1) cluster @@
  ...
-linesss- metadata:
-linesss-   annotations:
-linesss-     test: test-ann-val
-linesss-   creationTimestamp: "2006-01-02T15:04:05Z07:00"
-linesss-   generation: 1
  ...
-linesss- spec:
-linesss-   conversion:
-linesss-     strategy: None
-linesss-   group: kapp.example
-linesss-   names:
-linesss-     kind: Test
-linesss-     listKind: TestList
-linesss-     plural: tests
-linesss-     singular: test
-linesss-   scope: Namespaced
-linesss-   versions:`

	name := "test-default-config-exclude-diff-against-last-applied-status"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	var crd ClusterResource

	logger.Section("initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		crd = NewPresentClusterResource("customresourcedefinition", "tests.kapp.example", env.Namespace, kubectl)
		kind := crd.RawPath(ctlres.NewPathFromStrings([]string{"status", "acceptedNames", "kind"})).(string)

		require.Equal(t, "Test", kind)
	})

	logger.Section("second deploy (status field is ignored in diff)", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml)})

		require.Errorf(t, err, "Expected to receive error")

		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("deploy with some changes to a field", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "-c"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedYaml)})

		// status is ignored in diff
		checkChangesOutput(t, replaceTimestampWithDfaultValue(out), expectedChangesForUpdate)
	})

	logger.Section("deploy after some changes are made by external controller", func() {
		patchCRD := `[{ "op": "add", "path": "/metadata/annotations/test", "value": "test-ann-val"}]`

		// Patch CRD using kubectl so that smart diff is not used
		PatchClusterResource("crd", "tests.kapp.example", env.Namespace, patchCRD, kubectl)

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml)})

		// status is ignored when smart diff is not used
		checkChangesOutput(t, replaceTimestampWithDfaultValue(out), expectedChangesForExternalUpdate)
	})
}

func TestDefaultConfig_AggregatedClusterRole(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml := `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: servicebinding-aggregate
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      servicebinding.io/controller: "true"
rules: []
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    servicebinding.io/controller: "true"
  name: servicebinding-k8s-apps-workload
rules:
- apiGroups: [apps]
  resources: [daemonsets, deployments, replicasets, statefulsets]
  verbs: [get, list, watch, update, patch]
`

	name := "test-default-config-aggregate-cluster-role"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	var cr ClusterResource

	logger.Section("initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		cr = NewPresentClusterResource("clusterrole", "servicebinding-aggregate", env.Namespace, kubectl)
		ruleAPIGroup := cr.RawPath(ctlres.Path{
			ctlres.NewPathPartFromString("rules"),
			ctlres.NewPathPartFromIndex(0),
			ctlres.NewPathPartFromString("apiGroups"),
			ctlres.NewPathPartFromIndex(0),
		}).(string)

		require.Equal(t, "apps", ruleAPIGroup)
	})

	logger.Section("second deploy (rebase runs)", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml)})

		require.Errorf(t, err, "Expected to receive error")

		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})
}

func TestConfigHavingRegex(t *testing.T) {
	configMapResYaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-demo
  annotations:
    foo1: bar1
    foo2: bar2
data:
  player_initial_lives: "3"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-test
  annotations:
    foo2: bar2
data:
  player_initial_lives: "3"
`

	updatedConfigMapResYaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-demo
data:
  player_initial_lives: "3"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-test
data:
  player_initial_lives: "3"
`

	configYaml := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
rebaseRules:
  - path: [metadata, annotations, {regex: "^foo"}]
    type: %s
    sources: [new, existing]
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: ConfigMap}
`

	faultyConfigYaml := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
rebaseRules:
  - path: [metadata, annotations, {regex: }]
    type: %s
    sources: [new, existing]
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: ConfigMap}
`

	deploymentResYaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: default
  name: simple-app
spec:
  selector:
    matchLabels:
      simple-app: ""
  template:
    metadata:
      labels:
        simple-app: ""
    spec:
      containers:
      - name: simple-app
        image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
        env:
        - name: HELLO
          value: strange
        - name: HELLO_MSG
          value: stranger
`

	updatedDeploymentResYaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: default
  name: simple-app
spec:
  selector:
    matchLabels:
      simple-app: ""
  template:
    metadata:
      labels:
        simple-app: ""
    spec:
      containers:
      - name: simple-app
        image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
        env:
        - name: HELLO
        - name: HELLO_MSG
`

	deploymentConfig := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
  - path: [spec, template, spec, containers, {allIndexes: true}, env, %s]
    type: %s
    sources: [new, existing]
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
`

	deploymentConfigIndex := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
  - path: [spec, template, spec, containers, {allIndexes: true}, env, {index: 0}, %s]
    type: copy
    sources: [new, existing]
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
  - path: [spec, template, spec, containers, {allIndexes: true}, env, {index: 1}, %s]
    type: copy
    sources: [new, existing]
    resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
`

	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	fieldsExcludedInMatch := []string{"kapp.k14s.io/app", "creationTimestamp:", "resourceVersion:", "uid:", "selfLink:", "kapp.k14s.io/association"}
	name := "test-config-path-regex"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy configmaps with annotations", func() {
		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(configMapResYaml)})
	})

	logger.Section("deploy configmaps without annotations", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedConfigMapResYaml + fmt.Sprintf(configYaml, "copy"))})

		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("passing faulty config", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedConfigMapResYaml + fmt.Sprintf(faultyConfigYaml, "copy"))})

		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "panic: Unknown path part", "Expected to panic")
	})

	logger.Section("Remove all the annotation with remove config", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes-yaml"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(configMapResYaml + fmt.Sprintf(configYaml, "remove"))})

		expectedOutput := `
---
# update: configmap/game-demo (v1) namespace: kapp-test
apiVersion: v1
data:
  player_initial_lives: "3"
kind: ConfigMap
metadata:
  annotations: {}
  labels:
  name: game-demo
  namespace: kapp-test
---
# update: configmap/game-test (v1) namespace: kapp-test
apiVersion: v1
data:
  player_initial_lives: "3"
kind: ConfigMap
metadata:
  annotations: {}
  labels:
  name: game-test
  namespace: kapp-test
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(fieldsExcludedInMatch, out)

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")
	})

	logger.Section("Deployment resource", func() {
		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(deploymentResYaml)})
	})

	logger.Section("Deployment resource with remove value field and copying with rebase rule", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedDeploymentResYaml + fmt.Sprintf(deploymentConfig, "{allIndexes: true}, value", "copy"))})

		// no change as value field is copied again for all indexes in the updatedDeployment using config resource
		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("Deployment resource with remove value field and copying with rebase rule using regex", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedDeploymentResYaml + fmt.Sprintf(deploymentConfig, "{allIndexes: true}, {regex: \"^val\"}", "copy"))})

		// no change as value field is copied again using regex for all indexes in the updatedDeployment using config resource
		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("Deployment resource with remove value field and copying with rebase rule using index and field", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedDeploymentResYaml + fmt.Sprintf(deploymentConfigIndex, "value", "value"))})

		// no change as value field is copied again for both index of env 0 and 1 in the updatedDeployment using config resource
		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("Deployment resource with remove value field and copying with rebase rule using index and regex", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedDeploymentResYaml + fmt.Sprintf(deploymentConfigIndex, "{regex: \"^val\"}", "{regex: \"^val\"}"))})

		// no change as value field is copied again using regex for both index of env 0 and 1 in the updatedDeployment using config resource
		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
		require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")
	})

	logger.Section("Deployment resource with remove value field and unmatched regex", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedDeploymentResYaml + fmt.Sprintf(deploymentConfigIndex, "{regex: \"^tal\"}", "{regex: \"^tal\"}"))})

		// change exists as no field is present as per given regex and hence it was unable to copy the field
		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with pending changes (exit status 3)", "Expected to find stderr output")
	})

	logger.Section("Deployment resource with remove value field and unmatched regex and allIndex", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run", "--diff-exit-status"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(updatedDeploymentResYaml + fmt.Sprintf(deploymentConfig, "{allIndexes: true}, {regex: \"^tal\"}", "copy"))})

		// change exists as no field is present as per given regex on all the indexes and hence it was unable to copy the field
		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "Exiting after diffing with pending changes (exit status 3)", "Expected to find stderr output")
	})

}

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
