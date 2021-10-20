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

func TestAppsV1DaemonSetCreation(t *testing.T) {
	currentData := `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  generation: 1
`

	state := buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for generation 1 to be observed",
	}
	require.Equal(t, expectedState, state)

	currentData = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  generation: 1
status:
  desiredNumberScheduled: 3
  numberUnavailable: 1
  observedGeneration: 1
  updatedNumberScheduled: 2
`

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 updated pods to be scheduled",
	}
	require.Equal(t, expectedState, state)

	currentData = strings.Replace(currentData, "updatedNumberScheduled: 2", "updatedNumberScheduled: 3", -1)

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 unavailable pods",
	}
	require.Equal(t, expectedState, state)

	currentData = strings.Replace(currentData, "numberUnavailable: 1", "numberUnavailable: 0", -1)

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}

	require.Equal(t, expectedState, state)
}

func TestAppsV1DaemonSetUpdate(t *testing.T) {
	currentData := `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  generation: 2
status:
  desiredNumberScheduled: 3
  numberUnavailable: 0
  observedGeneration: 1
  updatedNumberScheduled: 3
`

	state := buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for generation 2 to be observed",
	}
	require.Equal(t, expectedState, state)

	// DaemonSet controller marks one of the "current" pods for deletion. (but all number unavailable is still 0, at this moment)
	currentData = strings.Replace(currentData, "updatedNumberScheduled: 3", "updatedNumberScheduled: 0", -1) // new image ==> new updateRevision ==> now, there are no pods of that revision
	currentData = strings.Replace(currentData, "observedGeneration: 1", "observedGeneration: 2", -1)

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 3 updated pods to be scheduled",
	}
	require.Equal(t, expectedState, state)

	// DaemonSet Controller deleted one pod, and replaced it with one updated pod.
	currentData = strings.Replace(currentData, "updatedNumberScheduled: 0", "updatedNumberScheduled: 1", -1)
	currentData = strings.Replace(currentData, "numberUnavailable: 0", "numberUnavailable: 1", -1)

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 2 updated pods to be scheduled",
	}
	require.Equal(t, expectedState, state)

	// DaemonSet Controller updated all pods, and all but the last pod are available.
	currentData = strings.Replace(currentData, "updatedNumberScheduled: 1", "updatedNumberScheduled: 3", -1)

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       false,
		Successful: false,
		Message:    "Waiting for 1 unavailable pods",
	}
	require.Equal(t, expectedState, state)

	currentData = strings.Replace(currentData, "numberUnavailable: 1", "numberUnavailable: 0", -1)

	state = buildDaemonSet(currentData, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: true,
		Message:    "",
	}
	require.Equal(t, expectedState, state)
}

func buildDaemonSet(resourcesBs string, t *testing.T) *ctlresm.AppsV1DaemonSet {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	require.NoErrorf(t, err, "Expected resources to parse")

	return ctlresm.NewAppsV1DaemonSet(newResources[0])
}
