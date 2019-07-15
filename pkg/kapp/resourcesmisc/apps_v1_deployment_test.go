package resourcesmisc_test

import (
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

func TestAppsV1DeploymentMinRepAvailable(t *testing.T) {
	configYAML := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  annotations:
    deployment.kubernetes.io/revision: "1"
    kapp.k14s.io/apps-v1-deployment-wait-minimum-replicas-available: "10%"
  generation: 1
spec:
  replicas: 50
status:
  observedGeneration: 1
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: app
  annotations:
    deployment.kubernetes.io/revision: "1"
  generation: 2
status:
  observedGeneration: 2
  replicas: 50
  availableReplicas: 4
`

	state := buildDep(configYAML, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for at least 1 available replicas (currently 4 available)",
	}
	if state != expectedState {
		t.Fatalf("Found incorrect state: %#v", state)
	}

	configYAML = strings.Replace(configYAML, "availableReplicas: 4", "availableReplicas: 5", -1)

	state = buildDep(configYAML, t).IsDoneApplying()
	if state != (ctlresm.DoneApplyState{Done: true, Successful: true, Message: ""}) {
		t.Fatalf("Found incorrect state: %#v", state)
	}
}

func buildDep(resourcesBs string, t *testing.T) *ctlresm.AppsV1Deployment {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	if err != nil {
		t.Fatalf("Expected resources to parse")
	}

	return ctlresm.NewAppsV1Deployment(newResources[0], newResources[1:])
}
