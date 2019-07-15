package diffgraph_test

import (
	"strings"
	"testing"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
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

	graph, err := buildChangeGraph(configYAML, t)
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

	_, err := buildChangeGraph(circularDep1YAML, t)
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle in grouped changes: job/job1 () cluster -> job/job2 () cluster -> job/job1 () cluster" {
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

	_, err := buildChangeGraph(circularDep2YAML, t)
	if err == nil {
		t.Fatalf("Expected graph to fail building")
	}
	if err.Error() != "Detected cycle in grouped changes: job/job1 () cluster -> job/job1 () cluster" {
		t.Fatalf("Expected to detect cycle: %s", err)
	}
}

func buildChangeGraph(resourcesBs string, t *testing.T) (*ctldgraph.ChangeGraph, error) {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	if err != nil {
		t.Fatalf("Expected resources to parse")
	}

	actualChanges := []ctldgraph.ActualChange{}
	for _, res := range newResources {
		actualChanges = append(actualChanges, actualChangeFromRes{res})
	}

	return ctldgraph.NewChangeGraph(actualChanges)
}

type actualChangeFromRes struct {
	res ctlres.Resource
}

func (a actualChangeFromRes) Resource() ctlres.Resource    { return a.res }
func (a actualChangeFromRes) Op() ctldgraph.ActualChangeOp { return ctldgraph.ActualChangeOpUpsert }
