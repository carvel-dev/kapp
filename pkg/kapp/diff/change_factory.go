// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type ChangeFactory struct {
	rebaseMods                               []ctlres.ResourceModWithMultiple
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod
	diffAgainstExistingFieldExclusionRules   []ctlres.FieldRemoveMod
	opts                                     ChangeOpts
}

type ChangeOpts struct {
	AllowAnchoredDiff bool
}

func NewChangeFactory(rebaseMods []ctlres.ResourceModWithMultiple,
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod, diffAgainstExistingFieldExclusionRules []ctlres.FieldRemoveMod, opts ChangeOpts) ChangeFactory {

	return ChangeFactory{rebaseMods, diffAgainstLastAppliedFieldExclusionMods, diffAgainstExistingFieldExclusionRules, opts}
}

func (f ChangeFactory) NewChangeAgainstLastApplied(existingRes, newRes ctlres.Resource) (Change, error) {
	// Retain original copy of existing resource and use it
	// for rebasing last applied resource and new resource.
	existingResForRebasing := existingRes

	if existingRes != nil {
		// If we have copy of last applied resource (assuming it's still "valid"),
		// use it as an existing resource to provide "smart" diff instead of
		// diffing against resource that is actually stored on cluster.
		lastAppliedRes := f.NewResourceWithHistory(existingRes).LastAppliedResource()
		if lastAppliedRes != nil {
			rebasedLastAppliedRes, err := NewRebasedResource(existingResForRebasing, lastAppliedRes, f.rebaseMods).Resource()
			if err != nil {
				return nil, err
			}
			existingRes = rebasedLastAppliedRes
		}

		historylessExistingRes, err := f.newResourceWithoutHistory(existingRes).Resource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}

	if newRes != nil {
		historylessNewRes, err := f.newResourceWithoutHistory(newRes).Resource()
		if err != nil {
			return nil, err
		}

		newRes = historylessNewRes
	}

	rebasedNewRes, err := NewRebasedResource(existingResForRebasing, newRes, f.rebaseMods).Resource()
	if err != nil {
		return nil, err
	}

	return NewChange(existingRes, rebasedNewRes, newRes, existingResForRebasing, f.opts), nil
}

func (f ChangeFactory) NewExactChange(existingRes, newRes ctlres.Resource) (Change, error) {
	if existingRes != nil {
		historylessExistingRes, err := f.newResourceWithoutHistory(existingRes).Resource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}

	if newRes != nil {
		historylessNewRes, err := f.newResourceWithoutHistory(newRes).Resource()
		if err != nil {
			return nil, err
		}

		newRes = historylessNewRes
	}

	rebasedNewRes, err := NewRebasedResource(existingRes, newRes, f.rebaseMods).Resource()
	if err != nil {
		return nil, err
	}

	return NewChange(existingRes, rebasedNewRes, newRes, existingRes, f.opts), nil
}

func (f ChangeFactory) NewResourceWithHistory(resource ctlres.Resource) ResourceWithHistory {
	return NewResourceWithHistory(resource, &f, f.diffAgainstLastAppliedFieldExclusionMods)
}

func (f ChangeFactory) newResourceWithoutHistory(resource ctlres.Resource) ResourceWithoutHistory {
	return NewResourceWithoutHistory(resource, f.diffAgainstExistingFieldExclusionRules)
}
