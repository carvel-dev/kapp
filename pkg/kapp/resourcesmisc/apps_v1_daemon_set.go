package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type Appsv1DaemonSet struct {
	resource ctlres.Resource
}

func NewAppsv1DaemonSet(resource ctlres.Resource) *Appsv1DaemonSet {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "DaemonSet",
	}
	if matcher.Matches(resource) {
		return &Appsv1DaemonSet{resource}
	}
	return nil
}

func (s Appsv1DaemonSet) IsDoneApplying() DoneApplyState {
	dset := appsv1.DaemonSet{}

	err := s.resource.AsTypedObj(&dset)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if dset.Generation != dset.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", dset.Generation)}
	}

	if dset.Status.NumberUnavailable > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for unavailable pods to go from %d to 0", dset.Status.NumberUnavailable)}
	}

	return DoneApplyState{Done: true, Successful: true}
}
