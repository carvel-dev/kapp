// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type APIRegistrationV1Beta1APIService struct {
	resource      ctlres.Resource
	ignoreFailing bool
}

func NewAPIRegistrationV1Beta1APIService(resource ctlres.Resource, ignoreFailing bool) *APIRegistrationV1Beta1APIService {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apiregistration.k8s.io/v1beta1",
		Kind:       "APIService",
	}
	if matcher.Matches(resource) {
		return &APIRegistrationV1Beta1APIService{resource, ignoreFailing}
	}
	return nil
}

func (s APIRegistrationV1Beta1APIService) IsDoneApplying() DoneApplyState {
	allTrue, msg := Conditions{s.resource}.IsSelectedTrue([]string{"Available"})

	if !allTrue && s.ignoreFailing {
		return DoneApplyState{Done: true, Successful: true, Message: fmt.Sprintf("Ignoring (%s)", msg)}
	}

	return DoneApplyState{Done: allTrue, Successful: allTrue, Message: msg}
}
