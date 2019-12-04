package resourcesmisc

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type APIRegistrationV1APIService struct {
	resource ctlres.Resource
}

func NewAPIRegistrationV1APIService(resource ctlres.Resource) *APIRegistrationV1APIService {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apiregistration.k8s.io/v1",
		Kind:       "APIService",
	}
	if matcher.Matches(resource) {
		return &APIRegistrationV1APIService{resource}
	}
	return nil
}

func (s APIRegistrationV1APIService) IsDoneApplying() DoneApplyState {
	allTrue, msg := Conditions{s.resource}.IsSelectedTrue([]string{"Available"})
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
