// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"reflect"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
		if !reflect.DeepEqual(firstData, map[string]interface{}{"keep": "", "delete": ""}) {
			t.Fatalf("Expected value to be correct: %#v", firstData)
		}

		secondData := NewPresentClusterResource("configmap", "second", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		if !reflect.DeepEqual(secondData, map[string]interface{}{"keep": "", "delete": ""}) {
			t.Fatalf("Expected value to be correct: %#v", secondData)
		}
	})

	logger.Section("check rebases", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		firstData := NewPresentClusterResource("configmap", "first", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		if !reflect.DeepEqual(firstData, map[string]interface{}{"keep": "", "keep2": ""}) {
			t.Fatalf("Expected value to be correct: %#v", firstData)
		}

		secondData := NewPresentClusterResource("configmap", "second", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"data"}))
		if !reflect.DeepEqual(secondData, map[string]interface{}{"keep": "", "keep2": "", "delete": ""}) {
			t.Fatalf("Expected value to be correct: %#v", secondData)
		}
	})
}

func TestYttRebaseRule(t *testing.T) {
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
    template.yml: |
      #@ load("@ytt:data", "data")
      #@ load("@ytt:overlay", "overlay")

      #@ res_name = data.values.existing.metadata.name
      #@ secrets = data.values.existing.secrets

      #@overlay/match by=overlay.all
      ---
      secrets:
      #@ for k in secrets:
      #@ if/end k.name.startswith(res_name+"-token-"):
      - name: #@ k.name
      #@ end
  resourceMatchers:
  - allMatcher: {}
`

	// ServiceAccount controller appends secret named '${metadata.name}-token-${rand}'
	yaml1 := `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa
secrets:
- name: some-secret`

	yaml2 := `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-sa
secrets:
- name: some-secret
- name: new-some-secret`

	name := "test-config-ytt-rebase"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	var generatedSecretName string

	logger.Section("initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(config + yaml1)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		if len(secrets) != 2 {
			t.Fatalf("Expected one set and one generated secret")
		}
		if !reflect.DeepEqual(secrets[0], map[string]interface{}{"name": "some-secret"}) {
			t.Fatalf("Expected provided secret at idx0: %#v", secrets[0])
		}
		generatedSecretName = secrets[1].(map[string]interface{})["name"].(string)
		if !strings.HasPrefix(generatedSecretName, "test-sa-token-") {
			t.Fatalf("Expected generated secret at idx1: %#v", secrets[1])
		}
	})

	logger.Section("deploy no change as rebase rule should retain generated secret", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(config + yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		expected := []map[string]string{}

		if !reflect.DeepEqual(resp.Tables[0].Rows, expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "Op:      0 create, 0 delete, 0 update, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})

	logger.Section("deploy with additional secret, but retain existing generated secret", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(config + yaml2)})

		secrets := NewPresentClusterResource("serviceaccount", "test-sa", env.Namespace, kubectl).RawPath(ctlres.NewPathFromStrings([]string{"secrets"})).([]interface{})
		if len(secrets) != 3 {
			t.Fatalf("Expected one set and one generated secret")
		}
		if !reflect.DeepEqual(secrets[0], map[string]interface{}{"name": "some-secret"}) {
			t.Fatalf("Expected provided secret at idx0: %#v", secrets[0])
		}
		if !reflect.DeepEqual(secrets[1], map[string]interface{}{"name": "new-some-secret"}) {
			t.Fatalf("Expected provided secret at idx1: %#v", secrets[0])
		}
		if !reflect.DeepEqual(secrets[2], map[string]interface{}{"name": generatedSecretName}) {
			t.Fatalf("Expected previous generated secret at idx2: %#v", secrets[1])
		}
	})
}
