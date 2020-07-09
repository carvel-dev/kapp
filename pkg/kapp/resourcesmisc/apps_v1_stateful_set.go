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
	fmt.Printf("Resource: %v\n", statefulSet)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	//updateStrategy := statefulSet.Spec.UpdateStrategy
	//if updateStrategy.Type == "OnDelete" {
	//	return DoneApplyState{Done: true, Successful: true}
	//}

	if statefulSet.Generation != statefulSet.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", statefulSet.Generation)}
	}

	// TODO: introduce logic for partition
	// Once ReadyReplicas == Replicas and UpdatedReplicas == Replicas: we can pass
	if statefulSet.Status.ReadyReplicas < statefulSet.Status.Replicas ||
		statefulSet.Status.UpdatedReplicas < statefulSet.Status.Replicas {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d replicas to be updated and ready (currently %d updated and %d ready)",
			statefulSet.Status.Replicas-min(statefulSet.Status.UpdatedReplicas, statefulSet.Status.ReadyReplicas),
			statefulSet.Status.UpdatedReplicas,
			statefulSet.Status.ReadyReplicas,
		)}
	}

	return DoneApplyState{Done: true, Successful: true}
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
