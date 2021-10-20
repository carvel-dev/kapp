// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph_test

import (
	"strings"
	"testing"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestChangeGraph(t *testing.T) {
	configYAML := `
kind: ConfigMap
metadata:
  name: app-config
  annotations: {}
---
kind: Job
metadata:
  name: import-etcd-into-db
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/import-etcd-into-db"
    kapp.k14s.io/change-rule: "upsert before deleting apps.big.co/etcd" # ref to removed object
---
kind: Job
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/db-migrations"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/import-etcd-into-db"
---
kind: Service
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
kind: Ingress
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
kind: Deployment
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/db-migrations"
---
kind: Job
metadata:
  name: app-health-check
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/deployment"
`

	graph, err := buildChangeGraph(configYAML, ctldgraph.ActualChangeOpUpsert, t)
	require.NoErrorf(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) configmap/app-config () cluster
(upsert) job/import-etcd-into-db () cluster
(upsert) job/migrations () cluster
  (upsert) job/import-etcd-into-db () cluster
(upsert) service/app () cluster
(upsert) ingress/app () cluster
(upsert) deployment/app () cluster
  (upsert) job/migrations () cluster
    (upsert) job/import-etcd-into-db () cluster
(upsert) job/app-health-check () cluster
  (upsert) service/app () cluster
  (upsert) ingress/app () cluster
  (upsert) deployment/app () cluster
    (upsert) job/migrations () cluster
      (upsert) job/import-etcd-into-db () cluster
`)
	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphWithConfDefaults(t *testing.T) {
	configYAML := `
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1
metadata:
  name: app-config
spec:
  group: app-group
  names:
    kind: app-kind
---
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
kind: Deployment
apiVersion: apps/v1
metadata:
  name: app
  namespace: app1
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
kind: Job
apiVersion: batch/v1
metadata:
  name: app-health-check
  namespace: app1
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/deployment"
`

	_, conf, err := ctlconf.NewConfFromResourcesWithDefaults(nil)
	require.NoErrorf(t, err, "Expected parsing conf defaults to succeed")

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
(upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
(upsert) namespace/app1 (v1) cluster
(upsert) configmap/app-config (v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
(upsert) deployment/app (apps/v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
  (upsert) configmap/app-config (v1) namespace: app1
    (upsert) namespace/app1 (v1) cluster
(upsert) job/app-health-check (batch/v1) namespace: app1
  (upsert) namespace/app1 (v1) cluster
  (upsert) deployment/app (apps/v1) namespace: app1
    (upsert) namespace/app1 (v1) cluster
    (upsert) configmap/app-config (v1) namespace: app1
      (upsert) namespace/app1 (v1) cluster
  (upsert) configmap/app-config (v1) namespace: app1
    (upsert) namespace/app1 (v1) cluster
`)

	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphWithMultipleRules(t *testing.T) {
	configYAML := `
kind: Job
metadata:
  name: import-etcd-into-db
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/import-etcd-into-db"
---
kind: Job
metadata:
  name: after-migrations
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/after-migrations"
---
kind: Job
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-rule.rule1: "upsert after upserting apps.big.co/import-etcd-into-db"
    kapp.k14s.io/change-rule.rule2: "upsert before upserting apps.big.co/after-migrations"
`

	graph, err := buildChangeGraph(configYAML, ctldgraph.ActualChangeOpUpsert, t)
	require.NoErrorf(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) job/import-etcd-into-db () cluster
(upsert) job/after-migrations () cluster
  (upsert) job/migrations () cluster
    (upsert) job/import-etcd-into-db () cluster
(upsert) job/migrations () cluster
  (upsert) job/import-etcd-into-db () cluster
`)

	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphWithMultipleGroups(t *testing.T) {
	configYAML := `
kind: Job
metadata:
  name: import-etcd-into-db
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/import-etcd-into-db"
---
kind: Job
metadata:
  name: after-migrations
  annotations:
    kapp.k14s.io/change-group.group1: "apps.big.co/after-migrations-1"
    kapp.k14s.io/change-group.group2: "apps.big.co/after-migrations-2"
---
kind: Job
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-rule.rule1: "upsert after upserting apps.big.co/import-etcd-into-db"
    kapp.k14s.io/change-rule.rule2: "upsert before upserting apps.big.co/after-migrations-1"
    kapp.k14s.io/change-rule.rule3: "upsert before upserting apps.big.co/after-migrations-2"
`

	graph, err := buildChangeGraph(configYAML, ctldgraph.ActualChangeOpUpsert, t)
	require.NoErrorf(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) job/import-etcd-into-db () cluster
(upsert) job/after-migrations () cluster
  (upsert) job/migrations () cluster
    (upsert) job/import-etcd-into-db () cluster
(upsert) job/migrations () cluster
  (upsert) job/import-etcd-into-db () cluster
`)

	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphWithDeletes(t *testing.T) {
	configYAML := `
kind: ConfigMap
metadata:
  name: app-config
  annotations: {}
---
kind: Job
metadata:
  name: import-etcd-into-db
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/import-etcd-into-db"
    kapp.k14s.io/change-rule: "upsert before deleting apps.big.co/etcd" # ref to removed object
---
kind: Job
metadata:
  name: migrations
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/db-migrations"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/import-etcd-into-db"
    kapp.k14s.io/change-rule.0: "delete before deleting apps.big.co/deployment"
---
kind: Service
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
kind: Ingress
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
---
kind: Deployment
metadata:
  name: app
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/deployment"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/db-migrations"
---
kind: Job
metadata:
  name: app-health-check
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/deployment"
    kapp.k14s.io/change-rule.0: "delete before deleting apps.big.co/db-migrations"
`

	graph, err := buildChangeGraph(configYAML, ctldgraph.ActualChangeOpDelete, t)
	require.NoErrorf(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(delete) configmap/app-config () cluster
(delete) job/import-etcd-into-db () cluster
(delete) job/migrations () cluster
  (delete) job/app-health-check () cluster
(delete) service/app () cluster
  (delete) job/migrations () cluster
    (delete) job/app-health-check () cluster
(delete) ingress/app () cluster
  (delete) job/migrations () cluster
    (delete) job/app-health-check () cluster
(delete) deployment/app () cluster
  (delete) job/migrations () cluster
    (delete) job/app-health-check () cluster
(delete) job/app-health-check () cluster
`)

	require.Equal(t, expectedOutput, output)
}

func TestChangeGraphCircularOther(t *testing.T) {
	circularDep1YAML := `
kind: Job
metadata:
  name: job1
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job1"
    kapp.k14s.io/change-rule: "upsert before upserting apps.big.co/job2"
---
kind: Job
metadata:
  name: job2
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job2"
    kapp.k14s.io/change-rule: "upsert before upserting apps.big.co/job1"
`

	_, err := buildChangeGraph(circularDep1YAML, ctldgraph.ActualChangeOpUpsert, t)
	require.Error(t, err, "Expected graph to fail building")

	expectedErr := "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)"
	require.EqualError(t, err, expectedErr, "Expected to detect cycle")
}

func TestChangeGraphCircularTransitive(t *testing.T) {
	circularDep1YAML := `
kind: Job
metadata:
  name: job1
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job1"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job3"
---
kind: Job
metadata:
  name: job2
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job2"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job1"
---
kind: Job
metadata:
  name: job3
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job3"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job2"
`

	_, err := buildChangeGraph(circularDep1YAML, ctldgraph.ActualChangeOpUpsert, t)
	require.Error(t, err, "Expected graph to fail building")

	expectedErr := "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job3 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)"
	require.EqualError(t, err, expectedErr, "Expected to detect cycle")
}

func TestChangeGraphCircularDirect(t *testing.T) {
	circularDep1YAML := `
kind: Job
metadata:
  name: job1
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job1"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job2"
---
kind: Job
metadata:
  name: job2
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job2"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job1"
`

	_, err := buildChangeGraph(circularDep1YAML, ctldgraph.ActualChangeOpUpsert, t)
	require.Error(t, err, "Expected graph to fail building")

	expectedErr := "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)"
	require.EqualError(t, err, expectedErr, "Expected to detect cycle")
}

func TestChangeGraphCircularWithinADep(t *testing.T) {
	circularDep1YAML := `
kind: Job
metadata:
  name: job3
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job1"
---
kind: Job
metadata:
  name: job1
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job1"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job2"
---
kind: Job
metadata:
  name: job2
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job2"
    kapp.k14s.io/change-rule: "upsert after upserting apps.big.co/job1"
`

	_, err := buildChangeGraph(circularDep1YAML, ctldgraph.ActualChangeOpUpsert, t)
	require.Error(t, err, "Expected graph to fail building")

	expectedErr := "Detected cycle while ordering changes: [job/job3 () cluster] -> [job/job1 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)"
	require.EqualError(t, err, expectedErr, "Expected to detect cycle")
}

func TestChangeGraphCircularSelf(t *testing.T) {
	circularDep2YAML := `
kind: Job
metadata:
  name: job1
  annotations:
    kapp.k14s.io/change-group: "apps.big.co/job1"
    kapp.k14s.io/change-rule: "upsert before upserting apps.big.co/job1"
`

	_, err := buildChangeGraph(circularDep2YAML, ctldgraph.ActualChangeOpUpsert, t)
	require.Error(t, err, "Expected graph to fail building")

	expectedErr := "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)"
	require.EqualError(t, err, expectedErr, "Expected to detect cycle")
}

func TestChangeGraphWithNamespaceAndCRDs(t *testing.T) {
	configYAML := `
apiVersion: v1
kind: Namespace
metadata:
  name: kapp-namespace-1
---
apiVersion: v1
kind: Namespace
metadata:
  name: kapp-namespace-2
---
apiVersion: v1
kind: Namespace
metadata:
  name: app
---
apiVersion: v1
kind: Secret
metadata:
  name: kapp-secret-1
  namespace: kapp-namespace-2
---
apiVersion: v1
kind: Secret
metadata:
  name: kapp-secret-2
  namespace: kapp-namespace-2
---
apiVersion: v1
kind: Secret
metadata:
  name: kapp-secret-3
  namespace: default
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1
metadata:
  name: kapp-crd-1
spec:
  group: appGroup
  names:
    kind: KappCRD1
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1
metadata:
  name: kapp-crd-2
spec:
  group: appGroup
  names:
    kind: KappCRD2
---
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1
metadata:
  name: kapp-crd-3
spec:
  group: appGroup
  names:
    kind: KappCRD3
---
kind: KappCRD1
apiVersion: appGroup/v1
metadata:
  name: kapp-cr-1
---
kind: KappCRD2
apiVersion: appGroup/v1
metadata:
  name: kapp-cr-2
`

	_, conf, err := ctlconf.NewConfFromResourcesWithDefaults(nil)
	require.NoError(t, err, "Expected parsing conf defaults to succeed")

	opts := buildGraphOpts{
		resourcesBs:         configYAML,
		op:                  ctldgraph.ActualChangeOpUpsert,
		changeGroupBindings: conf.ChangeGroupBindings(),
		changeRuleBindings:  conf.ChangeRuleBindings(),
	}

	graph, err := buildChangeGraphWithOpts(opts, t)
	require.NoError(t, err, "Expected graph to build")

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) namespace/kapp-namespace-1 (v1) cluster
(upsert) namespace/kapp-namespace-2 (v1) cluster
(upsert) namespace/app (v1) cluster
(upsert) secret/kapp-secret-1 (v1) namespace: kapp-namespace-2
  (upsert) namespace/kapp-namespace-2 (v1) cluster
(upsert) secret/kapp-secret-2 (v1) namespace: kapp-namespace-2
  (upsert) namespace/kapp-namespace-2 (v1) cluster
(upsert) secret/kapp-secret-3 (v1) namespace: default
(upsert) customresourcedefinition/kapp-crd-1 (apiextensions.k8s.io/v1) cluster
(upsert) customresourcedefinition/kapp-crd-2 (apiextensions.k8s.io/v1) cluster
(upsert) customresourcedefinition/kapp-crd-3 (apiextensions.k8s.io/v1) cluster
(upsert) kappcrd1/kapp-cr-1 (appGroup/v1) cluster
  (upsert) customresourcedefinition/kapp-crd-1 (apiextensions.k8s.io/v1) cluster
(upsert) kappcrd2/kapp-cr-2 (appGroup/v1) cluster
  (upsert) customresourcedefinition/kapp-crd-2 (apiextensions.k8s.io/v1) cluster
`)

	require.Equal(t, expectedOutput, output)
}

func TestGraphOrderWithClusterRoleAndClusterRoleBinding(t *testing.T) {
	configYAML := `
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-rbac-cluster-role
rules:
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: test-rbac-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: test-rbac-cluster-role
`

	_, conf, err := ctlconf.NewConfFromResourcesWithDefaults(nil)
	if err != nil {
		t.Fatalf("Error parsing conf defaults")
	}

	opts := buildGraphOpts{
		resourcesBs:         configYAML,
		op:                  ctldgraph.ActualChangeOpUpsert,
		changeGroupBindings: conf.ChangeGroupBindings(),
		changeRuleBindings:  conf.ChangeRuleBindings(),
	}

	graph, err := buildChangeGraphWithOpts(opts, t)
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) clusterrole/test-rbac-cluster-role (rbac.authorization.k8s.io/v1) cluster
(upsert) clusterrolebinding/test-rbac-cluster-role-binding (rbac.authorization.k8s.io/v1) cluster
  (upsert) clusterrole/test-rbac-cluster-role (rbac.authorization.k8s.io/v1) cluster
`)
	require.Equal(t, expectedOutput, output)
}

func buildChangeGraph(resourcesBs string, op ctldgraph.ActualChangeOp, t *testing.T) (*ctldgraph.ChangeGraph, error) {
	return buildChangeGraphWithOpts(buildGraphOpts{resourcesBs: resourcesBs, op: op}, t)
}

type buildGraphOpts struct {
	resources           []ctlres.Resource
	resourcesBs         string
	op                  ctldgraph.ActualChangeOp
	changeGroupBindings []ctlconf.ChangeGroupBinding
	changeRuleBindings  []ctlconf.ChangeRuleBinding
}

func buildChangeGraphWithOpts(opts buildGraphOpts, t *testing.T) (*ctldgraph.ChangeGraph, error) {
	var rs []ctlres.Resource

	if len(opts.resources) > 0 {
		rs = opts.resources
	} else {
		var err error
		rs, err = ctlres.NewFileResource(ctlres.NewBytesSource([]byte(opts.resourcesBs))).Resources()
		require.NoError(t, err, "Expected resources to parse")
	}

	actualChanges := []ctldgraph.ActualChange{}
	for _, res := range rs {
		actualChanges = append(actualChanges, actualChangeFromRes{res, opts.op})
	}

	return ctldgraph.NewChangeGraph(actualChanges,
		opts.changeGroupBindings, opts.changeRuleBindings, logger.NewTODOLogger())
}

type actualChangeFromRes struct {
	res ctlres.Resource
	op  ctldgraph.ActualChangeOp
}

func (a actualChangeFromRes) Resource() ctlres.Resource    { return a.res }
func (a actualChangeFromRes) Op() ctldgraph.ActualChangeOp { return a.op }
