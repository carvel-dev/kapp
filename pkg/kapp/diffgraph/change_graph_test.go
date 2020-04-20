package diffgraph_test

import (
	"strings"
	"testing"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

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

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", output, expectedOutput)
	}
}

func TestChangeGraphWithConfDefaults(t *testing.T) {
	configYAML := `
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1
metadata:
  name: app-config
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
	if err != nil {
		t.Fatalf("Error parsing conf defaults")
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
(upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
(upsert) namespace/app1 (v1) cluster
  (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
(upsert) configmap/app-config (v1) namespace: app1
  (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
  (upsert) namespace/app1 (v1) cluster
    (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
(upsert) deployment/app (apps/v1) namespace: app1
  (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
  (upsert) namespace/app1 (v1) cluster
    (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
  (upsert) configmap/app-config (v1) namespace: app1
    (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
    (upsert) namespace/app1 (v1) cluster
      (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
(upsert) job/app-health-check (batch/v1) namespace: app1
  (upsert) deployment/app (apps/v1) namespace: app1
    (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
    (upsert) namespace/app1 (v1) cluster
      (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
    (upsert) configmap/app-config (v1) namespace: app1
      (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
      (upsert) namespace/app1 (v1) cluster
        (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
  (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
  (upsert) namespace/app1 (v1) cluster
    (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
  (upsert) configmap/app-config (v1) namespace: app1
    (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
    (upsert) namespace/app1 (v1) cluster
      (upsert) customresourcedefinition/app-config (apiextensions.k8s.io/v1) cluster
`)

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", output, expectedOutput)
	}
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
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) job/import-etcd-into-db () cluster
(upsert) job/after-migrations () cluster
  (upsert) job/migrations () cluster
    (upsert) job/import-etcd-into-db () cluster
(upsert) job/migrations () cluster
  (upsert) job/import-etcd-into-db () cluster
`)

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", output, expectedOutput)
	}
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
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

	output := strings.TrimSpace(graph.PrintStr())
	expectedOutput := strings.TrimSpace(`
(upsert) job/import-etcd-into-db () cluster
(upsert) job/after-migrations () cluster
  (upsert) job/migrations () cluster
    (upsert) job/import-etcd-into-db () cluster
(upsert) job/migrations () cluster
  (upsert) job/import-etcd-into-db () cluster
`)

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", output, expectedOutput)
	}
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
	if err != nil {
		t.Fatalf("Expected graph to build")
	}

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

	if output != expectedOutput {
		t.Fatalf("Expected output to be >>>%s<<< but was >>>%s<<<", output, expectedOutput)
	}
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
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)" {
		t.Fatalf("Expected to detect cycle: %s", err)
	}
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
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job3 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)" {
		t.Fatalf("Expected to detect cycle: %s", err)
	}
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
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)" {
		t.Fatalf("Expected to detect cycle: %s", err)
	}
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
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle while ordering changes: [job/job3 () cluster] -> [job/job1 () cluster] -> [job/job2 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)" {
		t.Fatalf("Expected to detect cycle: %s", err)
	}
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
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle while ordering changes: [job/job1 () cluster] -> [job/job1 () cluster] (found repeated: job/job1 () cluster)" {
		t.Fatalf("Expected to detect cycle: %s", err)
	}
}

func buildChangeGraph(resourcesBs string, op ctldgraph.ActualChangeOp, t *testing.T) (*ctldgraph.ChangeGraph, error) {
	return buildChangeGraphWithOpts(buildGraphOpts{resourcesBs: resourcesBs, op: op}, t)
}

type buildGraphOpts struct {
	resources              []ctlres.Resource
	resourcesBs            string
	op                     ctldgraph.ActualChangeOp
	additionalChangeGroups []ctlconf.AdditionalChangeGroup
	additionalChangeRules  []ctlconf.AdditionalChangeRule
}

func buildChangeGraphWithOpts(opts buildGraphOpts, t *testing.T) (*ctldgraph.ChangeGraph, error) {
	var rs []ctlres.Resource

	if len(opts.resources) > 0 {
		rs = opts.resources
	} else {
		var err error
		rs, err = ctlres.NewFileResource(ctlres.NewBytesSource([]byte(opts.resourcesBs))).Resources()
		if err != nil {
			t.Fatalf("Expected resources to parse")
		}
	}

	actualChanges := []ctldgraph.ActualChange{}
	for _, res := range rs {
		actualChanges = append(actualChanges, actualChangeFromRes{res, opts.op})
	}

	return ctldgraph.NewChangeGraph(actualChanges,
		opts.additionalChangeGroups, opts.additionalChangeRules, logger.NewTODOLogger())
}

type actualChangeFromRes struct {
	res ctlres.Resource
	op  ctldgraph.ActualChangeOp
}

func (a actualChangeFromRes) Resource() ctlres.Resource    { return a.res }
func (a actualChangeFromRes) Op() ctldgraph.ActualChangeOp { return a.op }
