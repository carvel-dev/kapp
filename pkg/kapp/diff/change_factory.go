// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeFactory struct {
	rebaseMods                               []ctlres.ResourceModWithMultiple
	dropMods                                 []ctlres.ResourceModWithMultiple
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod
}

func NewChangeFactory(rebaseMods []ctlres.ResourceModWithMultiple, dropMods []ctlres.ResourceModWithMultiple,
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod) ChangeFactory {

	return ChangeFactory{rebaseMods, dropMods, diffAgainstLastAppliedFieldExclusionMods}
}

func (f ChangeFactory) NewChangeAgainstLastApplied(existingRes, newRes ctlres.Resource) (Change, error) {
	// Retain original copy of existing resource and use it
	// for rebasing last applied resource and new resource.
	existingResForRebasing := existingRes
	tmpRmod := f.rebaseMods
	var rMod []ctlres.ResourceModWithMultiple
	for _, mod := range f.rebaseMods {
		if m, ok := mod.(ctlres.FieldRemoveMod); ok && m.Path.AsString() == "status" {
			fmt.Printf("Mod status: [%+v], %s\n", m, m.Path.AsString())
			continue
		}
		rMod = append(rMod, mod)
	}
	f.rebaseMods = rMod
	if existingRes != nil {
		// If we have copy of last applied resource (assuming it's still "valid"),
		// use it as an existing resource to provide "smart" diff instead of
		// diffing against resource that is actually stored on cluster.
		lastAppliedRes := f.NewResourceWithHistory(existingRes).LastAppliedResource()
		if lastAppliedRes != nil {
			rebasedLastAppliedRes, err := NewRebasedResource(existingResForRebasing, lastAppliedRes, rMod).Resource()
			if err != nil {
				return nil, err
			}
			existingRes = rebasedLastAppliedRes
		}

		historylessExistingRes, err := f.NewResourceWithHistory(existingRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}
	f.rebaseMods = tmpRmod
	if newRes != nil {
		historylessNewRes, err := f.NewResourceWithHistory(newRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		newRes = historylessNewRes
	}

	rebasedNewRes, err := NewRebasedResource(existingResForRebasing, newRes, f.rebaseMods).Resource()
	if err != nil {
		return nil, err
	}

	rebasedExistingRes, err := NewRebasedResource(rebasedNewRes, existingRes, f.rebaseMods).Resource()
	if err != nil {
		return nil, err
	}
	// rebasedExistingRes, rebasedNewRes, err := f.NewResourceWithDroppedFields(existingRes, rebasedNewRes)
	// if err != nil {
	// 	return nil, err
	// }

	return NewChange(existingRes, rebasedNewRes, newRes, rebasedExistingRes), nil
}

func (f ChangeFactory) NewResourceWithDroppedFields(existingRes, newRes ctlres.Resource) (ctlres.Resource,
	ctlres.Resource, error) {

	newRes, err := NewRebasedResource(existingRes, newRes, f.dropMods).Resource()
	if err != nil {
		return nil, nil, err
	}

	if existingRes == nil {
		return existingRes, newRes, nil
	}

	existingRes, err = NewRebasedResource(newRes, existingRes, f.dropMods).Resource()
	if err != nil {
		return nil, nil, err
	}

	return existingRes, newRes, nil
}

func (f ChangeFactory) NewExactChange(existingRes, newRes ctlres.Resource) (Change, error) {
	if existingRes != nil {
		historylessExistingRes, err := f.NewResourceWithHistory(existingRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}

	if newRes != nil {
		historylessNewRes, err := f.NewResourceWithHistory(newRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		newRes = historylessNewRes
	}

	rebasedNewRes, err := NewRebasedResource(existingRes, newRes, f.rebaseMods).Resource()
	if err != nil {
		return nil, err
	}

	return NewChange(existingRes, rebasedNewRes, newRes, existingRes), nil
}

func (f ChangeFactory) NewResourceWithHistory(resource ctlres.Resource) ResourceWithHistory {
	return NewResourceWithHistory(resource, &f, f.diffAgainstLastAppliedFieldExclusionMods)
}
