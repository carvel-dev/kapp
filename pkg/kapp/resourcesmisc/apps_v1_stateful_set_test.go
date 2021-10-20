// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc_test

import (
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"github.com/stretchr/testify/require"
)

func TestAppsV1StatefulSetCreation(t *testing.T) {
	currentData := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 1
spec:
  replicas: 3
`

	state := buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for generation 1 to be observed",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	currentData = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 1
spec:
  replicas: 3
status:
  replicas: 1
  currentReplicas: 1
  observedGeneration: 1
  updatedReplicas: 1
  readyReplicas: 0
`

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 2 replicas to be updated",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	currentData = strings.Replace(currentData, "currentReplicas: 1", "currentReplicas: 3", -1)
	currentData = strings.Replace(currentData, "updatedReplicas: 1", "updatedReplicas: 3", -1)
	currentData = strings.Replace(currentData, "readyReplicas: 0", "readyReplicas: 2", -1)
	currentData = strings.Replace(currentData, "status:\n  replicas: 1", "status:\n  replicas: 3", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be ready",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	currentData = strings.Replace(currentData, "readyReplicas: 2", "readyReplicas: 3", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

}

func TestAppsV1StatefulSetUpdate(t *testing.T) {
	currentData := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 2
spec:
  replicas: 3
status:
  replicas: 3
  currentReplicas: 3
  observedGeneration: 1
  updatedReplicas: 3
  readyReplicas: 3
`

	state := buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for generation 2 to be observed",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet controller marks one of the "current" pods for deletion. (but all 3 are still ready, at this moment)
	currentData = strings.Replace(currentData, "updatedReplicas: 3", "updatedReplicas: 0", -1) // new image ==> new updateRevision ==> now, there are no pods of that revision
	currentData = strings.Replace(currentData, "currentReplicas: 3", "currentReplicas: 2", -1)
	currentData = strings.Replace(currentData, "observedGeneration: 1", "observedGeneration: 2", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 3 replicas to be updated",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet Controller deleted one pod, and replaced it with one updated pod.
	currentData = strings.Replace(currentData, "readyReplicas: 3", "readyReplicas: 2", -1)
	currentData = strings.Replace(currentData, "updatedReplicas: 0", "updatedReplicas: 1", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 2 replicas to be updated",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet Controller updated all pods, and all but the last pod are ready.
	currentData = strings.Replace(currentData, "updatedReplicas: 1", "updatedReplicas: 3", -1)
	currentData = strings.Replace(currentData, "currentReplicas: 2", "currentReplicas: 0", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be ready",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	currentData = strings.Replace(currentData, "readyReplicas: 2", "readyReplicas: 3", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

}

func TestAppsV1StatefulSetUpdatePartition(t *testing.T) {
	currentData := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 2
spec:
  replicas: 3
  updateStrategy:
    rollingUpdate:
      partition: 1
status:
  replicas: 3
  currentReplicas: 3
  observedGeneration: 1
  updatedReplicas: 3
  readyReplicas: 3
`

	state := buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for generation 2 to be observed",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet controller marks one of the "current" pods for deletion. (but all 3 are still ready, at this moment)
	currentData = strings.Replace(currentData, "updatedReplicas: 3", "updatedReplicas: 0", -1) // new image ==> new updateRevision ==> now, there are no pods of that revision
	currentData = strings.Replace(currentData, "currentReplicas: 3", "currentReplicas: 2", -1)
	currentData = strings.Replace(currentData, "observedGeneration: 1", "observedGeneration: 2", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 2 replicas to be updated (updating only 2 of 3 total)",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet Controller deleted one pod, and replaced it with one updated pod.
	currentData = strings.Replace(currentData, "readyReplicas: 3", "readyReplicas: 2", -1)
	currentData = strings.Replace(currentData, "updatedReplicas: 0", "updatedReplicas: 1", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be updated (updating only 2 of 3 total)",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet Controller updated all pods, and all but the last pod are ready.
	currentData = strings.Replace(currentData, "updatedReplicas: 1", "updatedReplicas: 2", -1)
	currentData = strings.Replace(currentData, "currentReplicas: 2", "currentReplicas: 1", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be ready",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	currentData = strings.Replace(currentData, "readyReplicas: 2", "readyReplicas: 3", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

}

func TestAppsV1StatefulSetScaleDown(t *testing.T) {
	currentData := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  generation: 2
spec:
  replicas: 1
status:
  replicas: 2
  currentReplicas: 2
  observedGeneration: 1
  updatedReplicas: 2
  readyReplicas: 2
`

	state := buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for generation 2 to be observed",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet controller marks one of the "current" pods for deletion. Updated == current since scaling change does not create a new revision.
	currentData = strings.Replace(currentData, "updatedReplicas: 2", "updatedReplicas: 1", -1)
	currentData = strings.Replace(currentData, "currentReplicas: 2", "currentReplicas: 1", -1)
	currentData = strings.Replace(currentData, "observedGeneration: 1", "observedGeneration: 2", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be deleted",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	currentData = strings.Replace(currentData, "readyReplicas: 2", "readyReplicas: 1", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 replicas to be deleted",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")

	// StatefulSet Controller has finished removing replicas
	currentData = strings.Replace(currentData, "status:\n  replicas: 2", "status:\n  replicas: 1", -1)

	state = buildStatefulSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	require.Equal(t, expectedState, state, "Found incorrect state")
}

func buildStatefulSet(resourcesBs string, t *testing.T) *ctlresm.AppsV1StatefulSet {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	require.NoErrorf(t, err, "Expected resources to parse")

	return ctlresm.NewAppsV1StatefulSet(newResources[0], nil)
}
