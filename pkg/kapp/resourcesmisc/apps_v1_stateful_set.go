// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type AppsV1StatefulSet struct {
	resource ctlres.Resource
}

func NewAppsV1StatefulSet(resource ctlres.Resource, _ []ctlres.Resource) *AppsV1StatefulSet {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "StatefulSet",
	}
	if matcher.Matches(resource) {
		return &AppsV1StatefulSet{resource}
	}
	return nil
}

func (s AppsV1StatefulSet) IsDoneApplying() DoneApplyState {
	statefulSet := appsv1.StatefulSet{}

	err := s.resource.AsTypedObj(&statefulSet)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if statefulSet.Generation != statefulSet.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", statefulSet.Generation)}
	}

	if statefulSet.Spec.Replicas == nil {
		return DoneApplyState{Done: true, Successful: false,
			Message: fmt.Sprintf("Error: Failed to find spec.replicas")}
	}

	toUpdate := *statefulSet.Spec.Replicas
	clarification := ""
	if s.partition(statefulSet) {
		toUpdate -= *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition
		clarification = fmt.Sprintf(" (updating only %d of %d total)",
			toUpdate, *statefulSet.Spec.Replicas)
	}

	// ensure replicas have been updated
	notUpdated := toUpdate - statefulSet.Status.UpdatedReplicas
	if notUpdated > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d replicas to be updated%s", notUpdated, clarification)}
	}

	// ensure replicas are available
	notReady := *statefulSet.Spec.Replicas - statefulSet.Status.ReadyReplicas
	if notReady > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d replicas to be ready", notReady)}
	}

	// ensure all replicas have been deleted when scaling down
	notDeleted := statefulSet.Status.Replicas - *statefulSet.Spec.Replicas
	if notDeleted > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d replicas to be deleted", notDeleted)}
	}

	return DoneApplyState{Done: true, Successful: true}
}

func (AppsV1StatefulSet) partition(statefulSet appsv1.StatefulSet) bool {
	return statefulSet.Spec.UpdateStrategy.RollingUpdate != nil &&
		statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil &&
		*statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition > 0
}
