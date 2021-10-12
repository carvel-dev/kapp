// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph_test

import (
	"strings"
	"testing"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestChangeGraphWithAdditionalOrderRules(t *testing.T) {
	configYAML := `
kind: Namespace
apiVersion: v1
metadata:
  name: app1
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: app-config
  namespace: app1
`

	confYAML := `
kind: Config
apiVersion: kapp.k14s.io/v1alpha1

changeGroupBindings:
- name: test.kapp.k14s.io/namespace
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Namespace, apiVersion: v1}

changeRuleBindings:
- rules:
  - "upsert after upserting test.kapp.k14s.io/namespace"
  resourceMatchers:
  - apiVersionKindMatcher: {kind: ConfigMap, apiVersion: v1}
`

	_, conf, err := ctlconf.NewConfFromResources([]ctlres.Resource{ctlres.MustNewResourceFromBytes([]byte(confYAML))})
	require.NoErrorf(t, err, "Expected parsing conf to succeed")

	opts := buildGraphOpts{
		resourcesBs:         configYAML,
		op:                  ctldgraph.ActualChangeOpUpsert,
		changeGroupBindings: conf.ChangeGroupBindings(),
		changeRuleBindings:  conf.ChangeRuleBindings(),
	}

	graph, err := buildChangeGraphWithOpts(opts, t)
	require.NoErrorf(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) namespace/app1 (v1) cluster
(upsert) configmap/app-config (v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
`)

	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphWithOptionalRulesThatProduceCycles(t *testing.T) {
	configYAML := `
kind: Namespace
apiVersion: v1
metadata:
  name: app1
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: app-config
  namespace: app1
---
kind: Secret
apiVersion: v1
metadata:
  name: app-config
  namespace: app1
`

	confYAML := `
kind: Config
apiVersion: kapp.k14s.io/v1alpha1

changeGroupBindings:
- name: test.kapp.k14s.io/namespace
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Namespace, apiVersion: v1}
- name: test.kapp.k14s.io/configmap
  resourceMatchers:
  - apiVersionKindMatcher: {kind: ConfigMap, apiVersion: v1}
- name: test.kapp.k14s.io/secret
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Secret, apiVersion: v1}

changeRuleBindings:
- rules:
  - "upsert after upserting test.kapp.k14s.io/configmap"
  ignoreIfCyclical: true
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Secret, apiVersion: v1}

- rules:
  - "upsert after upserting test.kapp.k14s.io/secret"
  ignoreIfCyclical: true
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Namespace, apiVersion: v1}

- rules:
  - "upsert after upserting test.kapp.k14s.io/namespace"
  resourceMatchers:
  - apiVersionKindMatcher: {kind: ConfigMap, apiVersion: v1}
`

	_, conf, err := ctlconf.NewConfFromResources([]ctlres.Resource{ctlres.MustNewResourceFromBytes([]byte(confYAML))})
	require.NoErrorf(t, err, "Expected parsing conf to succeed")

	opts := buildGraphOpts{
		resourcesBs:         configYAML,
		op:                  ctldgraph.ActualChangeOpUpsert,
		changeGroupBindings: conf.ChangeGroupBindings(),
		changeRuleBindings:  conf.ChangeRuleBindings(),
	}

	graph, err := buildChangeGraphWithOpts(opts, t)
	require.NoErrorf(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) namespace/app1 (v1) cluster
  (upsert) secret/app-config (v1) namespace: app1
(upsert) configmap/app-config (v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
    (upsert) secret/app-config (v1) namespace: app1
(upsert) secret/app-config (v1) namespace: app1
`)

	require.Equal(t, expectedOutput, output)
}
