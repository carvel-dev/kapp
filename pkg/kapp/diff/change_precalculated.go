package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangePrecalculated struct {
	existingRes, newRes ctlres.Resource

	// appliedRes is an unmodified copy of what's being applied
	appliedRes ctlres.Resource

	op       ChangeOp
	textDiff TextDiff
	opsDiff  OpsDiff
}

var _ Change = &ChangePrecalculated{}

func NewChangePrecalculated(existingRes, newRes, appliedRes ctlres.Resource) *ChangePrecalculated {
	if existingRes == nil && newRes == nil {
		panic("Expected either existingRes or newRes be non-nil")
	}

	if existingRes != nil {
		existingRes = existingRes.DeepCopy()
	}
	if newRes != nil {
		newRes = newRes.DeepCopy()
	}
	if appliedRes != nil {
		appliedRes = appliedRes.DeepCopy()
	}

	return &ChangePrecalculated{existingRes: existingRes, newRes: newRes, appliedRes: appliedRes}
}

func (d *ChangePrecalculated) NewOrExistingResource() ctlres.Resource {
	if d.newRes != nil {
		return d.newRes
	}
	if d.existingRes != nil {
		return d.existingRes
	}
	panic("Not possible")
}

func (d *ChangePrecalculated) NewResource() ctlres.Resource      { return d.newRes }
func (d *ChangePrecalculated) ExistingResource() ctlres.Resource { return d.existingRes }
func (d *ChangePrecalculated) AppliedResource() ctlres.Resource  { return d.appliedRes }

func (d *ChangePrecalculated) Op() ChangeOp       { return d.op }
func (d *ChangePrecalculated) TextDiff() TextDiff { return d.textDiff }
func (d *ChangePrecalculated) OpsDiff() OpsDiff   { return d.opsDiff }

func (d *ChangePrecalculated) IsIgnored() bool       { return false }
func (d *ChangePrecalculated) IgnoredReason() string { return "" }
