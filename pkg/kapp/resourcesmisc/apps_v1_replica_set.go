// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type AppsV1ReplicaSet struct {
	resource ctlres.Resource
}

func NewAppsV1ReplicaSet(resource ctlres.Resource) *AppsV1ReplicaSet {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "ReplicaSet",
	}
	if matcher.Matches(resource) {
		return &AppsV1ReplicaSet{resource}
	}
	return nil
}

func (s AppsV1ReplicaSet) IsDoneApplying() DoneApplyState {
	rs := appsv1.ReplicaSet{}

	err := s.resource.AsTypedObj(&rs)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if rs.Generation != rs.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", rs.Generation)}
	}

	if rs.Status.Replicas != rs.Status.AvailableReplicas {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d unavailable replicas", rs.Status.Replicas-rs.Status.AvailableReplicas)}
	}

	return DoneApplyState{Done: true, Successful: true}
}
