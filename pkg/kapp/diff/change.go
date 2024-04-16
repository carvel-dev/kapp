// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-patch/patch"
)

type ChangeOp string

const (
	ChangeOpAdd    ChangeOp = "add"
	ChangeOpDelete ChangeOp = "delete"
	ChangeOpUpdate ChangeOp = "update"
	ChangeOpKeep   ChangeOp = "keep" // unchanged
	ChangeOpExists ChangeOp = "exists"
	ChangeOpNoop   ChangeOp = "noop"
)

type Change interface {
	NewOrExistingResource() ctlres.Resource
	NewResource() ctlres.Resource
	ExistingResource() ctlres.Resource
	AppliedResource() ctlres.Resource
	ClusterOriginalResource() ctlres.Resource

	Op() ChangeOp
	ConfigurableTextDiff() *ConfigurableTextDiff
	OpsDiff() OpsDiff

	IsIgnored() bool
}

type ChangeImpl struct {
	existingRes, newRes ctlres.Resource

	// appliedRes is an unmodified copy of what's being applied
	appliedRes ctlres.Resource

	// clusterOriginalRes is an unmodified copy of what's present on the cluster
	clusterOriginalRes ctlres.Resource

	configurableTextDiff *ConfigurableTextDiff
	opsDiff              *OpsDiff
	changeOpVal          ChangeOp

	opts ChangeOpts
}

var _ Change = &ChangeImpl{}

func NewChange(existingRes, newRes, appliedRes, clusterOriginalRes ctlres.Resource, opts ChangeOpts) *ChangeImpl {
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
	if clusterOriginalRes != nil {
		clusterOriginalRes = clusterOriginalRes.DeepCopy()
	}

	return &ChangeImpl{existingRes: existingRes, newRes: newRes, appliedRes: appliedRes, clusterOriginalRes: clusterOriginalRes, opts: opts}
}

func (d *ChangeImpl) NewOrExistingResource() ctlres.Resource {
	if d.newRes != nil {
		return d.newRes
	}
	if d.existingRes != nil {
		return d.existingRes
	}
	panic("Not possible")
}

func (d *ChangeImpl) NewResource() ctlres.Resource             { return d.newRes }
func (d *ChangeImpl) ExistingResource() ctlres.Resource        { return d.existingRes }
func (d *ChangeImpl) AppliedResource() ctlres.Resource         { return d.appliedRes }
func (d *ChangeImpl) ClusterOriginalResource() ctlres.Resource { return d.clusterOriginalRes }

func (d *ChangeImpl) Op() ChangeOp {
	if d.changeOpVal == "" {
		d.changeOpVal = d.op()
	}
	return d.changeOpVal
}

func (d *ChangeImpl) op() ChangeOp {
	if d.newRes != nil {
		if _, hasNoopAnnotation := d.newRes.Annotations()[ctlres.NoopAnnKey]; hasNoopAnnotation {
			return ChangeOpNoop
		}
	}

	if d.existingRes == nil {
		if d.newResHasExistsAnnotation() {
			return ChangeOpExists
		}
		return ChangeOpAdd
	}

	if d.newRes == nil {
		return ChangeOpDelete
	}

	if d.ConfigurableTextDiff().Full().HasChanges() {
		if d.newResHasExistsAnnotation() {
			return ChangeOpKeep
		}
		return ChangeOpUpdate
	}

	return ChangeOpKeep
}

func (d *ChangeImpl) IsIgnored() bool { return d.isIgnoredTransient() }

func (d *ChangeImpl) isIgnoredTransient() bool {
	return d.existingRes != nil && d.newRes == nil && d.existingRes.Transient()
}

func (d *ChangeImpl) ConfigurableTextDiff() *ConfigurableTextDiff {
	// diff is called very often, so memoize
	if d.configurableTextDiff == nil {
		d.configurableTextDiff = NewConfigurableTextDiff(d.existingRes, d.newRes, d.IsIgnored(), d.opts)
	}
	return d.configurableTextDiff
}

func (d *ChangeImpl) OpsDiff() OpsDiff {
	if d.opsDiff != nil {
		return *d.opsDiff
	}

	opsDiff := d.calculateOpsDiff()
	d.opsDiff = &opsDiff

	return *d.opsDiff
}

func (d *ChangeImpl) calculateOpsDiff() OpsDiff {
	return OpsDiff(patch.Diff{Left: d.existingRes.UnstructuredObject(), Right: d.newRes.UnstructuredObject()}.Calculate())
}

func (d *ChangeImpl) newResHasExistsAnnotation() bool {
	_, hasExistsAnnotation := d.newRes.Annotations()[ctlres.ExistsAnnKey]
	return hasExistsAnnotation
}
