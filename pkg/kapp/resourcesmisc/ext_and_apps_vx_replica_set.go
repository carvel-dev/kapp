// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type ExtensionsAndAppsVxReplicaSet struct {
	resource ctlres.Resource
}

func NewExtensionsAndAppsVxReplicaSet(resource ctlres.Resource) *ExtensionsAndAppsVxReplicaSet {
	extMatcher := ctlres.APIGroupKindMatcher{
		APIGroup: "extensions",
		Kind:     "ReplicaSet",
	}
	appsMatcher := ctlres.APIGroupKindMatcher{
		APIGroup: "apps",
		Kind:     "ReplicaSet",
	}
	if extMatcher.Matches(resource) || appsMatcher.Matches(resource) {
		return &ExtensionsAndAppsVxReplicaSet{resource}
	}
	return nil
}

func (s ExtensionsAndAppsVxReplicaSet) IsDoneApplyingWithMinimum(minAvailable int) DoneApplyState {
	rs := appsv1.ReplicaSet{}

	// TODO unsafely unmarshals any replica set version
	err := s.resource.AsUncheckedTypedObj(&rs)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if rs.Generation != rs.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", rs.Generation)}
	}

	if int(rs.Status.AvailableReplicas) < minAvailable {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for at least %d available replicas (currently %d available)",
			minAvailable-int(rs.Status.AvailableReplicas), rs.Status.AvailableReplicas)}
	}

	return DoneApplyState{Done: true, Successful: true}
}
