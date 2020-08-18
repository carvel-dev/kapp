// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"reflect"
	"strings"
	"testing"

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
