package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type RebasedResource struct {
	existingRes, newRes ctlres.Resource
	mods                []ctlres.FieldCopyMod
}

func NewRebasedResource(existingRes, newRes ctlres.Resource, mods []ctlres.FieldCopyMod) RebasedResource {
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

	if r.existingRes == nil {
		return result, nil // all done rebasing
	}

	resSources := map[ctlres.FieldCopyModSource]ctlres.Resource{
		ctlres.FieldCopyModSourceNew:      r.newRes.DeepCopy(), // cant be newRes as it will be modified in place
		ctlres.FieldCopyModSourceExisting: r.existingRes,
	}

	for _, t := range r.mods {
		err := t.ApplyFromMultiple(result, resSources)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
