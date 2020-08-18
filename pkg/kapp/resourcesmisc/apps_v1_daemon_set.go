// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
)

type AppsV1DaemonSet struct {
	resource ctlres.Resource
}

func NewAppsV1DaemonSet(resource ctlres.Resource) *AppsV1DaemonSet {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "DaemonSet",
	}
	if matcher.Matches(resource) {
		return &AppsV1DaemonSet{resource}
	}
	return nil
}

func (s AppsV1DaemonSet) IsDoneApplying() DoneApplyState {
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
			"Waiting for %d unavailable pods", dset.Status.NumberUnavailable)}
	}

	return DoneApplyState{Done: true, Successful: true}
}
