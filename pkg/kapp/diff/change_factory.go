// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"context"
	"fmt"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/types"
)

type ChangeFactory struct {
	rebaseMods                               []ctlres.ResourceModWithMultiple
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod
	resources                                ctlres.Resources
}

type ChangeFactoryFunc func(ctx context.Context, existingRes, newRes ctlres.Resource) (Change, error)

var _ ChangeFactoryFunc = ChangeFactory{}.NewChangeSSA
var _ ChangeFactoryFunc = ChangeFactory{}.NewChangeAgainstLastApplied
var _ ChangeFactoryFunc = ChangeFactory{}.NewExactChange

func NewChangeFactory(rebaseMods []ctlres.ResourceModWithMultiple,
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod, resources ctlres.Resources) ChangeFactory {

	return ChangeFactory{rebaseMods, diffAgainstLastAppliedFieldExclusionMods, resources}
}

func (f ChangeFactory) NewChangeSSA(ctx context.Context, existingRes, newRes ctlres.Resource) (Change, error) {
	if existingRes != nil {
		historylessExistingRes, err := f.NewResourceWithHistory(existingRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}

	newResAsIs := newRes
	if newResAsIs != nil {
		historylessNewRes, err := f.NewResourceWithHistory(newRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		newResAsIs = historylessNewRes
	}

	dryRunRes := newResAsIs
	if dryRunRes != nil && existingRes != nil {
		newResBytes, _ := newRes.AsYAMLBytes()
		dryRunResult, err := f.resources.Patch(existingRes, types.ApplyPatchType, newResBytes, true)
		if err != nil {
			return nil, fmt.Errorf("SSA dry run: %s", err)
		}
		dryRunRes = dryRunResult
	}

	return NewChangeSSA(existingRes, newResAsIs, dryRunRes), nil
}

func (f ChangeFactory) NewChangeAgainstLastApplied(ctx context.Context, existingRes, newRes ctlres.Resource) (Change, error) {
	// Retain original copy of existing resource and use it
	// for rebasing last applied resource and new resource.
	existingResForRebasing := existingRes
	var err error

	if existingRes != nil {
		// Strip rebasing "base" object of kapp history so that it never shows up in the diff, regardless
		// of rebase rules used
		existingResForRebasing, err = f.NewResourceWithHistory(existingRes).HistorylessResource()
		if err != nil {
			return nil, err
		}
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
	}

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

	return NewChange(existingRes, rebasedNewRes, newRes), nil
}

func (f ChangeFactory) NewExactChange(ctx context.Context, existingRes, newRes ctlres.Resource) (Change, error) {
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

	return NewChange(existingRes, rebasedNewRes, newRes), nil
}

func (f ChangeFactory) NewResourceWithHistory(resource ctlres.Resource) ResourceWithHistory {
	return NewResourceWithHistory(resource, &f, f.diffAgainstLastAppliedFieldExclusionMods)
}
