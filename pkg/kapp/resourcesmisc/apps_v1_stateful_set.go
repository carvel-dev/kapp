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

	if partition(statefulSet) {
		if statefulSet.Status.UpdatedReplicas + *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition <
			*statefulSet.Spec.Replicas {
			return DoneApplyState{Done: false, Message: fmt.Sprintf(
				"Waiting for replicas [%d-%d] to be updated due to partition update strategy (currently %d/%d updated)",
				*statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition+1,
				*statefulSet.Spec.Replicas,
				statefulSet.Status.UpdatedReplicas,
				*statefulSet.Spec.Replicas-*statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition)}
		}

		if statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas {
			return DoneApplyState{Done: false, Message: fmt.Sprintf(
				"Waiting for replicas [%d-%d] to be ready due to partition update strategy (currently %d/%d ready)",
				*statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition+1,
				*statefulSet.Spec.Replicas,
				statefulSet.Status.ReadyReplicas,
				*statefulSet.Spec.Replicas)}
		}
	} else {
		// Once ReadyReplicas == Replicas and UpdatedReplicas == Replicas: we can pass
		if statefulSet.Status.UpdatedReplicas < *statefulSet.Spec.Replicas {
			return DoneApplyState{Done: false, Message: fmt.Sprintf(
				"Waiting for %d replicas to be updated (currently %d updated)",
				*statefulSet.Spec.Replicas,
				statefulSet.Status.UpdatedReplicas)}
		}

		if statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas {
			return DoneApplyState{Done: false, Message: fmt.Sprintf(
				"Waiting for %d replicas to be ready (currently %d ready)",
				*statefulSet.Spec.Replicas,
				statefulSet.Status.ReadyReplicas)}
		}
	}

	return DoneApplyState{Done: true, Successful: true}
}

func partition(statefulSet appsv1.StatefulSet) bool {
	// todo: nil check
	return true
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
