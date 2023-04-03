// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"

	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

type RebasedResource struct {
	existingRes, newRes ctlres.Resource
	mods                []ctlres.ResourceModWithMultiple
}

func NewRebasedResource(existingRes, newRes ctlres.Resource, mods []ctlres.ResourceModWithMultiple) RebasedResource {
	if existingRes == nil && newRes == nil {
		panic("Expected either existingRes or newRes be non-nil")
	}

	if existingRes != nil {
		existingRes = existingRes.DeepCopy()
	}
	if newRes != nil {
		newRes = newRes.DeepCopy()
	}

	return RebasedResource{existingRes: existingRes, newRes: newRes, mods: mods}
}

func (r RebasedResource) Resource() (ctlres.Resource, error) {
	if r.newRes == nil {
		return nil, nil // nothing to rebase
	}

	result := r.newRes.DeepCopy()
	resultDesc := result.Description() // capture since resource could change

	if r.existingRes == nil {
		return result, nil // all done rebasing
	}

	for _, t := range r.mods {
		if t.IsResourceMatching(result) {
			// copy newRes and existingRes as they may be modified in place
			resSources := map[ctlres.FieldCopyModSource]ctlres.Resource{
				ctlres.FieldCopyModSourceNew:      r.newRes.DeepCopy(),
				ctlres.FieldCopyModSourceExisting: r.existingRes.DeepCopy(),
				// Might be useful for more advanced rebase rules like ytt-based
				ctlres.FieldCopyModSource("_current"): result.DeepCopy(),
			}

			err := t.ApplyFromMultiple(result, resSources)
			if err != nil {
				return nil, fmt.Errorf("Applying rebase rule to %s: %w", resultDesc, err)
			}
		}
	}

	return result, nil
}
