package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type Deploymentv1 struct {
	resource ctlres.Resource
}

func NewDeploymentv1(resource ctlres.Resource) *Deploymentv1 {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}
	if matcher.Matches(resource) {
		return &Deploymentv1{resource}
	}
	return nil
}

func (s Deploymentv1) IsDoneApplying() DoneApplyState {
	dep := appsv1.Deployment{}

	err := s.resource.AsTypedObj(&dep)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if dep.Generation != dep.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", dep.Generation)}
	}

	if dep.Status.UnavailableReplicas > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for unavailable replicas to go from %d to 0", dep.Status.UnavailableReplicas)}
	}

	return DoneApplyState{Done: true, Successful: true}
}
