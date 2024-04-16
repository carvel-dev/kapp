// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type AppsV1DaemonSet struct {
	resource ctlres.Resource
}

func NewAppsV1DaemonSet(resource ctlres.Resource) *AppsV1DaemonSet {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "DaemonSet",
	}
	if matcher.Matches(resource) {
		return &AppsV1DaemonSet{resource}
	}
	return nil
}

func (s AppsV1DaemonSet) IsDoneApplying() DoneApplyState {
	dset := appsv1.DaemonSet{}

	err := s.resource.AsTypedObj(&dset)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if dset.Generation != dset.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", dset.Generation)}
	}

	// ensure updated pods are actually scheduled before checking number unavailable to avoid
	// race condition between pod scheduler and kapp state check
	notReady := dset.Status.DesiredNumberScheduled - dset.Status.UpdatedNumberScheduled
	if notReady > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d updated pods to be scheduled", notReady)}
	}

	if dset.Status.NumberUnavailable > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d unavailable pods", dset.Status.NumberUnavailable)}
	}

	return DoneApplyState{Done: true, Successful: true}
}
