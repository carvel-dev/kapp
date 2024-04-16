// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc_test

import (
	"strings"
	"testing"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, expectedState, state, "Found incorrect state")

	configYAML = strings.Replace(configYAML, "availableReplicas: 4", "availableReplicas: 5", -1)

	state = buildDep(configYAML, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")
}

func TestAppsV1DeploymentReplicaFailure(t *testing.T) {

	depYAML := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  generation: 1
status:
  observedGeneration: 1
  conditions:
  - type: Progressing
    status: "False"
    reason: "ProgressDeadlineExceeded"
    message: "Progress deadline exceeded"
`

	deployment := buildDep(depYAML, t)
	state := deployment.IsDoneApplying()

	expectedState := ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    "Deployment is not progressing: ProgressDeadlineExceeded (message: Progress deadline exceeded)",
	}

	require.Equal(t, expectedState, state, "Found incorrect state")

	depYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  generation: 1
status:
  observedGeneration: 1
  conditions:
  - type: ReplicaFailure
    status: "True"
    reason: "FailedCreate"
    message: "Failed to create pods"
`
	deployment = buildDep(depYAML, t)
	state = deployment.IsDoneApplying()

	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Deployment has encountered replica failure: FailedCreate (message: Failed to create pods)",
	}

	require.Equal(t, expectedState, state, "Found incorrect state")
}

func buildDep(resourcesBs string, t *testing.T) *ctlresm.AppsV1Deployment {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	require.NoErrorf(t, err, "Expected resources to parse")

	return ctlresm.NewAppsV1Deployment(newResources[0], newResources[1:])
}
