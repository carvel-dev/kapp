package diffgraph_test

import (
	"strings"
	"testing"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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

additionalChangeGroups:
- name: test.kapp.k14s.io/namespace
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Namespace, apiVersion: v1}

additionalChangeRules:
- rules:
  - "upsert after upserting test.kapp.k14s.io/namespace"
  resourceMatchers:
  - apiVersionKindMatcher: {kind: ConfigMap, apiVersion: v1}
`

	_, conf, err := ctlconf.NewConfFromResources([]ctlres.Resource{ctlres.MustNewResourceFromBytes([]byte(confYAML))})
	if err != nil {
		t.Fatalf("Expected parsing conf to succeed")
	}

	opts := buildGraphOpts{
		resourcesBs:            configYAML,
		op:                     ctldgraph.ActualChangeOpUpsert,
		additionalChangeGroups: conf.AdditionalChangeGroups(),
		additionalChangeRules:  conf.AdditionalChangeRules(),
	}

	graph, err := buildChangeGraphWithOpts(opts, t)
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) namespace/app1 (v1) cluster
(upsert) configmap/app-config (v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
`)

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", expectedOutput, output)
	}
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

additionalChangeGroups:
- name: test.kapp.k14s.io/namespace
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Namespace, apiVersion: v1}
- name: test.kapp.k14s.io/configmap
  resourceMatchers:
  - apiVersionKindMatcher: {kind: ConfigMap, apiVersion: v1}
- name: test.kapp.k14s.io/secret
  resourceMatchers:
  - apiVersionKindMatcher: {kind: Secret, apiVersion: v1}

additionalChangeRules:
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
	if err != nil {
		t.Fatalf("Expected parsing conf to succeed")
	}

	opts := buildGraphOpts{
		resourcesBs:            configYAML,
		op:                     ctldgraph.ActualChangeOpUpsert,
		additionalChangeGroups: conf.AdditionalChangeGroups(),
		additionalChangeRules:  conf.AdditionalChangeRules(),
	}

	graph, err := buildChangeGraphWithOpts(opts, t)
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) namespace/app1 (v1) cluster
  (upsert) secret/app-config (v1) namespace: app1
(upsert) configmap/app-config (v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
    (upsert) secret/app-config (v1) namespace: app1
(upsert) secret/app-config (v1) namespace: app1
`)

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", expectedOutput, output)
	}
}
