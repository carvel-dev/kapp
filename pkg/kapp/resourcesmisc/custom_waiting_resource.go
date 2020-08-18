// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CustomWaitingResource struct {
	resource ctlres.Resource
	waitRule ctlconf.WaitRule
}

func NewCustomWaitingResource(resource ctlres.Resource, waitRules []ctlconf.WaitRule) *CustomWaitingResource {
	for _, rule := range waitRules {
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
	Type    string
	Status  string
	Reason  string
	Message string
}

func (s CustomWaitingResource) IsDoneApplying() DoneApplyState {
	obj := customWaitingResourceStruct{}

	err := s.resource.AsUncheckedTypedObj(&obj)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
			"Error: Failed obj conversion: %s", err)}
	}

	if s.waitRule.SupportsObservedGeneration && obj.Metadata.Generation != obj.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", obj.Metadata.Generation)}
	}

	// Check on failure conditions first
	for _, condMatcher := range s.waitRule.ConditionMatchers {
		for _, cond := range obj.Status.Conditions {
			if cond.Type == condMatcher.Type && cond.Status == condMatcher.Status {
				if condMatcher.Failure {
					return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
						"Encountered failure condition %s == %s: %s (message: %s)",
						cond.Type, condMatcher.Status, cond.Reason, cond.Message)}
				}
			}
		}
	}

	// If no failure conditions found, check on successful ones
	for _, condMatcher := range s.waitRule.ConditionMatchers {
		for _, cond := range obj.Status.Conditions {
			if cond.Type == condMatcher.Type && cond.Status == condMatcher.Status {
				if condMatcher.Success {
					return DoneApplyState{Done: true, Successful: true, Message: fmt.Sprintf(
						"Encountered successful condition %s == %s: %s (message: %s)",
						cond.Type, condMatcher.Status, cond.Reason, cond.Message)}
				}
			}
		}
	}

	return DoneApplyState{Done: false, Message: "No failing or successful conditions found"}
}
