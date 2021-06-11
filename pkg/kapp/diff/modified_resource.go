// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ModifiedResource struct {
	existingRes, newRes ctlres.Resource
	mods                []ctlres.ResourceModWithMultiple
}

func NewModifiedResource(existingRes, newRes ctlres.Resource, mods []ctlres.ResourceModWithMultiple) ModifiedResource {
	if existingRes == nil && newRes == nil {
		panic("Expected either existingRes or newRes be non-nil")
	}

	if existingRes != nil {
		existingRes = existingRes.DeepCopy()
	}
	if newRes != nil {
		newRes = newRes.DeepCopy()
	}

	return ModifiedResource{existingRes: existingRes, newRes: newRes, mods: mods}
}

func (r ModifiedResource) Resource() (ctlres.Resource, error) {
	if r.newRes == nil {
		return nil, nil // nothing to modify
	}

	result := r.newRes.DeepCopy()

	for _, t := range r.mods {
		var existingRes ctlres.Resource
		if r.existingRes != nil {
			existingRes = r.existingRes.DeepCopy()
		}

		// copy newRes and existingRes as they may be be modified in place
		resSources := map[ctlres.FieldCopyModSource]ctlres.Resource{
			ctlres.FieldCopyModSourceNew: r.newRes.DeepCopy(),
			// key is always present, but may be nil
			ctlres.FieldCopyModSourceExisting: existingRes,
		}

		err := t.ApplyFromMultiple(result, resSources)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
