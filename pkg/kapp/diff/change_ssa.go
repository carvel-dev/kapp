// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"github.com/cppforlife/go-patch/patch"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"gopkg.in/yaml.v2"
)

// Like ChangeImpl, but generates diff against dryRunRes
type ChangeSSA struct {
	existingRes, newRes ctlres.Resource

	// result off SSA of newRes on top of existingRes. Used to
	// calculate diff
	dryRunRes ctlres.Resource

	configurableTextDiff *ConfigurableTextDiff
	opsDiff              *OpsDiff
}

var _ Change = &ChangeSSA{}

func NewChangeSSA(existingRes, newRes, dryRunRes ctlres.Resource) *ChangeSSA {
	if existingRes == nil && newRes == nil {
		panic("Expected either existingRes or newRes be non-nil")
	}

	if existingRes != nil {
		existingRes = existingRes.DeepCopy()
	}
	if newRes != nil {
		newRes = newRes.DeepCopy()
	}

	if dryRunRes != nil {
		dryRunRes = dryRunRes.DeepCopy()
	}

	return &ChangeSSA{existingRes: existingRes, newRes: newRes, dryRunRes: dryRunRes}
}

func (d *ChangeSSA) NewOrExistingResource() ctlres.Resource {
	if d.newRes != nil {
		return d.newRes
	}
	if d.existingRes != nil {
		return d.existingRes
	}
	panic("Not possible")
}

func (d *ChangeSSA) NewResource() ctlres.Resource      { return d.newRes }
func (d *ChangeSSA) ExistingResource() ctlres.Resource { return d.existingRes }
func (d *ChangeSSA) AppliedResource() ctlres.Resource  { return d.newRes }

func (d *ChangeSSA) Op() ChangeOp {
	if d.newRes != nil {
		if _, hasNoopAnnotation := d.newRes.Annotations()[NoopAnnKey]; hasNoopAnnotation {
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

func (d *ChangeSSA) IsIgnored() bool { return d.isIgnoredTransient() }

func (d *ChangeSSA) isIgnoredTransient() bool {
	return d.existingRes != nil && d.newRes == nil && d.existingRes.Transient()
}

func (d *ChangeSSA) ConfigurableTextDiff() *ConfigurableTextDiff {
	// diff is called very often, so memoize
	if d.configurableTextDiff == nil {
		d.configurableTextDiff = NewConfigurableTextDiff(d.existingRes, d.dryRunRes, d.IsIgnored())
	}
	return d.configurableTextDiff
}

func (d *ChangeSSA) OpsDiff() OpsDiff {
	if d.opsDiff != nil {
		return *d.opsDiff
	}

	opsDiff := d.calculateOpsDiff()
	d.opsDiff = &opsDiff

	return *d.opsDiff
}

func (d *ChangeSSA) calculateOpsDiff() OpsDiff {
	var existingObj interface{}
	var newObj interface{}

	if d.existingRes != nil {
		existingBytes, err := d.existingRes.AsYAMLBytes()
		if err != nil {
			panic("yamling existingRes") // TODO panic
		}

		err = yaml.Unmarshal(existingBytes, &existingObj)
		if err != nil {
			panic("unyamling existingRes") // TODO panic
		}
	}

	if d.dryRunRes != nil {
		newBytes, err := d.dryRunRes.AsYAMLBytes()
		if err != nil {
			panic("yamling newRes") // TODO panic
		}

		err = yaml.Unmarshal(newBytes, &newObj)
		if err != nil {
			panic("unyamling newRes") // TODO panic
		}
	} else if d.IsIgnored() {
		newObj = existingObj // show as no changes
	}

	return OpsDiff(patch.Diff{Left: existingObj, Right: newObj}.Calculate())
}

func (d *ChangeSSA) newResHasExistsAnnotation() bool {
	_, hasExistsAnnotation := d.newRes.Annotations()[ExistsAnnKey]
	return hasExistsAnnotation
}
