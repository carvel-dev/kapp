// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type ChangePrecalculated struct {
	existingRes, newRes ctlres.Resource

	// appliedRes is an unmodified copy of what's being applied
	appliedRes ctlres.Resource

	op                   ChangeOp
	configurableTextDiff *ConfigurableTextDiff
	opsDiff              OpsDiff
}

var _ Change = &ChangePrecalculated{}

func NewChangePrecalculated(existingRes, newRes, appliedRes ctlres.Resource,
	op ChangeOp, configurableTextDiff *ConfigurableTextDiff, opsDiff OpsDiff) *ChangePrecalculated {

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

	return &ChangePrecalculated{existingRes: existingRes, newRes: newRes, appliedRes: appliedRes,
		op: op, configurableTextDiff: configurableTextDiff, opsDiff: opsDiff}
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

func (d *ChangePrecalculated) NewResource() ctlres.Resource             { return d.newRes }
func (d *ChangePrecalculated) ExistingResource() ctlres.Resource        { return d.existingRes }
func (d *ChangePrecalculated) AppliedResource() ctlres.Resource         { return d.appliedRes }
func (d *ChangePrecalculated) ClusterOriginalResource() ctlres.Resource { return d.existingRes }

func (d *ChangePrecalculated) Op() ChangeOp { return d.op }
func (d *ChangePrecalculated) ConfigurableTextDiff() *ConfigurableTextDiff {
	return d.configurableTextDiff
}
func (d *ChangePrecalculated) OpsDiff() OpsDiff { return d.opsDiff }

func (d *ChangePrecalculated) IsIgnored() bool { return false }
