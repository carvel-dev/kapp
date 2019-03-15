package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
)

type Servicev1 struct {
	resource ctlres.Resource
}

func NewServicev1(resource ctlres.Resource) *Servicev1 {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "v1",
		Kind:       "Service",
	}
	if matcher.Matches(resource) {
		return &Servicev1{resource}
	}
	return nil
}

func (s Servicev1) IsDoneApplying() DoneApplyState {
	svc := corev1.Service{}

	err := s.resource.AsTypedObj(&svc)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if svc.Spec.Type == corev1.ServiceTypeExternalName {
		return DoneApplyState{Done: true, Successful: true, Message: "External service"}
	}

	if svc.Spec.ClusterIP != corev1.ClusterIPNone && len(svc.Spec.ClusterIP) == 0 {
		return DoneApplyState{Done: false, Message: "ClusterIP is empty"}
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return DoneApplyState{Done: false, Message: "Load balancer ingress is empty"}
		}
	}

	return DoneApplyState{Done: true, Successful: true}
}
