package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const (
	reconcileStateAnnKey = "kapp.k14s.io/reconcile-state" // values: "ok", "fail", "ongoing"
	reconcileInfoAnnKey  = "kapp.k14s.io/reconcile-info"  // values: "", "msg"
)

type Reconciling struct {
	resource ctlres.Resource
}

func NewReconciling(resource ctlres.Resource) *Reconciling {
	if _, found := resource.Annotations()[reconcileStateAnnKey]; found {
		return &Reconciling{resource}
	}
	return nil
}

func (s Reconciling) IsDoneApplying() DoneApplyState {
	info := s.resource.Annotations()[reconcileInfoAnnKey]

	switch s.resource.Annotations()[reconcileStateAnnKey] {
	case "ok":
		return DoneApplyState{Done: true, Successful: true, Message: info}
	case "fail":
		return DoneApplyState{Done: true, Successful: false, Message: info}
	case "ongoing":
		return DoneApplyState{Done: false, Message: info}
	default:
		return DoneApplyState{Done: true, Successful: false,
			Message: fmt.Sprintf("Error: unknown reconcile state")}
	}
}
