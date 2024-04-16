// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type APIRegistrationV1APIService struct {
	resource      ctlres.Resource
	ignoreFailing bool
}

func NewAPIRegistrationV1APIService(resource ctlres.Resource, ignoreFailing bool) *APIRegistrationV1APIService {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apiregistration.k8s.io/v1",
		Kind:       "APIService",
	}
	if matcher.Matches(resource) {
		return &APIRegistrationV1APIService{resource, ignoreFailing}
	}
	return nil
}

func (s APIRegistrationV1APIService) IsDoneApplying() DoneApplyState {
	allTrue, msg := Conditions{s.resource}.IsSelectedTrue([]string{"Available"})

	if !allTrue && s.ignoreFailing {
		return DoneApplyState{Done: true, Successful: true, Message: fmt.Sprintf("Ignoring (%s)", msg)}
	}

	return DoneApplyState{Done: allTrue, Successful: allTrue, Message: msg}
}

/*

status:
  conditions:
  - lastTransitionTime: 2019-12-03T16:52:14Z
    message: all checks passed
    reason: Passed
    status: "True"
    type: Available

*/
