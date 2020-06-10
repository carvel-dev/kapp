package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CustomResource struct {
	resource    ctlres.Resource
	waitingRule WaitingRuleMod
}

func NewCustomResource(resource ctlres.Resource) *CustomResource {
	waitingRule, isMatch := getWaitingRule(resource)
	if isMatch {
		return &CustomResource{resource, waitingRule}
	}
	return nil
}

type genericResource struct {
	Metadata metav1.ObjectMeta
	Status   struct {
		ObservedGeneration int64
		Conditions         []genericCondition
	}
}

// genericCondition describes the generic condition fields
type genericCondition struct {
	// Type of condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, or Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

func (s CustomResource) IsDoneApplying() DoneApplyState {
	wr := s.waitingRule
	obj := genericResource{}
	err := s.resource.AsUncheckedTypedObj(&obj)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Unexpected error unmarshalling genericCondition on resource %v. Err: %v", s.resource.Name(), err)}
	}
	if wr.SupportsObservedGeneration && obj.Metadata.Generation != obj.Status.ObservedGeneration {
		return DoneApplyState{Done: false}
	}
	for _, fc := range wr.FailureConditions {
		for _, cond := range obj.Status.Conditions {
			if cond.Type == fc && cond.Status == corev1.ConditionTrue {
				return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Apply failed with failure condition: %v", cond.Message)}
			}
		}
	}
	for _, sc := range wr.SuccessfulConditions {
		for _, cond := range obj.Status.Conditions {
			if cond.Type == sc && cond.Status != corev1.ConditionTrue {
				return DoneApplyState{Done: false, Successful: false, Message: cond.Message}
			}
		}
	}

	return DoneApplyState{Done: true, Successful: true}
}

var globalWaitingRules []WaitingRuleMod

// We keep track of the config's custom waiting rules globally in this package to avoid
// propogating this as function parameters across the codebase
func SetGlobalWaitingRules(rules []WaitingRuleMod) {
	globalWaitingRules = rules
}

type WaitingRuleMod struct {
	SupportsObservedGeneration bool                     `json:"supportsObservedGeneration"`
	SuccessfulConditions       []string                 `json:"successfulConditions"`
	FailureConditions          []string                 `json:"failureConditions"`
	ResourceMatchers           []ctlres.ResourceMatcher `json:"resourceMatchers"`
}

// Find waiting rule for specified resource
func getWaitingRule(res ctlres.Resource) (WaitingRuleMod, bool) {
	rules := globalWaitingRules
	isMatch := false
	mod := WaitingRuleMod{}
	for _, rule := range rules {
		for _, matcher := range rule.ResourceMatchers {
			if matcher.Matches(res) {
				isMatch = true
				mod.SupportsObservedGeneration = rule.SupportsObservedGeneration
				mod.SuccessfulConditions = append(mod.SuccessfulConditions, rule.SuccessfulConditions...)
				mod.FailureConditions = append(mod.FailureConditions, rule.FailureConditions...)
			}
		}
	}
	return mod, isMatch
}
