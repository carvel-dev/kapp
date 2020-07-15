package resourcesmisc

import (
	"fmt"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type AppsV1StatefulSet struct {
	resource ctlres.Resource
}

func NewAppsV1StatefulSet(resource ctlres.Resource, associatedRs []ctlres.Resource) *AppsV1StatefulSet {
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

	replicasToUpdate := *statefulSet.Spec.Replicas
	updateString := ""
	if partition(statefulSet) {
		replicasToUpdate -= *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition
		updateString = fmt.Sprintf(" (of %d total)", *statefulSet.Spec.Replicas)
	}

	if statefulSet.Status.UpdatedReplicas  < replicasToUpdate {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d%s replicas to be updated (currently %d of %d)",
			replicasToUpdate,
			updateString,
			statefulSet.Status.UpdatedReplicas,
			replicasToUpdate)}
	}


	if statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d replicas to be ready (currently %d of %d)",
			*statefulSet.Spec.Replicas,
			statefulSet.Status.ReadyReplicas,
			*statefulSet.Spec.Replicas)}
	}

	return DoneApplyState{Done: true, Successful: true}
}

func partition(statefulSet appsv1.StatefulSet) bool {
	return statefulSet.Spec.UpdateStrategy.RollingUpdate != nil &&
		*statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition > 0
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
