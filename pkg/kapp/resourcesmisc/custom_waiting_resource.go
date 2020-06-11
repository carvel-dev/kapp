package resourcesmisc

import (
	"fmt"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CustomWaitingResource struct {
	resource    ctlres.Resource
	waitingRule ctlconf.WaitingRule
}

func NewCustomWaitingResource(resource ctlres.Resource, waitingRules []ctlconf.WaitingRule) *CustomWaitingResource {
	for _, rule := range waitingRules {
		if rule.ResourceMatcher().Matches(resource) {
			return &CustomWaitingResource{resource, rule}
		}
	}
	return nil
}

type customWaitingResourceStruct struct {
	Metadata metav1.ObjectMeta
	Status   struct {
		ObservedGeneration int64
		Conditions         []customWaitingResourceCondition
	}
}

type customWaitingResourceCondition struct {
	Type    string                 `json:"type"`
	Status  corev1.ConditionStatus `json:"status"`
	Reason  string                 `json:"reason,omitempty"`
	Message string                 `json:"message,omitempty"`
}

func (s CustomWaitingResource) IsDoneApplying() DoneApplyState {
	obj := customWaitingResourceStruct{}

	err := s.resource.AsUncheckedTypedObj(&obj)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
			"Error: Failed obj conversion: %s", err)}
	}

	if s.waitingRule.SupportsObservedGeneration && obj.Metadata.Generation != obj.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", obj.Metadata.Generation)}
	}

	for _, fc := range s.waitingRule.FailureConditions {
		for _, cond := range obj.Status.Conditions {
			if cond.Type == fc && cond.Status == corev1.ConditionTrue {
				return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
					"Encountered failure condition %s: %s (message: %s)", cond.Type, cond.Reason, cond.Message)}
			}
		}
	}

	for _, sc := range s.waitingRule.SuccessfulConditions {
		for _, cond := range obj.Status.Conditions {
			if cond.Type == sc && cond.Status == corev1.ConditionTrue {
				return DoneApplyState{Done: true, Successful: true, Message: fmt.Sprintf(
					"Encountered successful condition %s: %s (message: %s)", cond.Type, cond.Reason, cond.Message)}
			}
		}
	}

	return DoneApplyState{Done: false, Message: "No failing or successful conditions found"}
}
