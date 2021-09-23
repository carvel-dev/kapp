// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"strings"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeSetOpts struct {
	AgainstLastApplied bool
}

type ChangeSet struct {
	existingRs, newRs []ctlres.Resource
	opts              ChangeSetOpts
	changeFactory     ChangeFactory
}

func NewChangeSet(existingRs, newRs []ctlres.Resource,
	opts ChangeSetOpts, changeFactory ChangeFactory) *ChangeSet {

	return &ChangeSet{existingRs, newRs, opts, changeFactory}
}

func (d ChangeSet) Calculate() ([]Change, error) {
	changeFactoryFunc := d.changeFactory.NewExactChange
	if d.opts.AgainstLastApplied {
		changeFactoryFunc = d.changeFactory.NewChangeAgainstLastApplied
	}

	existingRsMap := map[string]ctlres.Resource{}
	alreadyChecked := map[string]struct{}{}
	changes := []Change{}

	for _, existingRes := range d.existingRs {
		existingRsMap[ctlres.NewUniqueResourceKey(existingRes).String()] = existingRes
	}

	// Go through new set of resources and compare to existing set of resources
	for _, newRes := range d.newRs {
		newRes := newRes
		newResKey := ctlres.NewUniqueResourceKey(newRes).String()

		var change Change
		var err error

		if existingRes, found := existingRsMap[newResKey]; found {
			change, err = changeFactoryFunc(existingRes, newRes)
			if err != nil {
				return nil, err
			}
		} else {
			change, err = changeFactoryFunc(nil, newRes)
			if err != nil {
				return nil, err
			}
		}

		changes = append(changes, change)
		alreadyChecked[newResKey] = struct{}{}
	}

	// Find existing resources that were not already diffed (not in new set of resources)
	for _, existingRes := range d.existingRs {
		existingRes := existingRes
		existingResKey := ctlres.NewUniqueResourceKey(existingRes).String()

		if _, found := alreadyChecked[existingResKey]; !found {
			change, err := changeFactoryFunc(existingRes, nil)
			if err != nil {
				return nil, err
			}

			changes = append(changes, change)
			alreadyChecked[existingResKey] = struct{}{}
		}
	}

	return d.collapseChangesWithSameUID(changes)
}

func (d ChangeSet) collapseChangesWithSameUID(changes []Change) ([]Change, error) {
	changeIdxsByUID := map[string][]int{}

	for i, change := range changes {
		// New resources do not have a UID assigned yet
		if change.ExistingResource() != nil {
			uid := change.ExistingResource().UID()
			if len(uid) > 0 {
				changeIdxsByUID[uid] = append(changeIdxsByUID[uid], i)
			}
		}
	}

	for _, idxs := range changeIdxsByUID {
		// One change per UID is typical case
		if len(idxs) == 1 {
			continue
		}

		var changeDescs []string
		for _, idx := range idxs {
			changeDescs = append(changeDescs, changes[idx].NewOrExistingResource().Description())
		}

		// When there are multiple UID matches it means
		// that resource api group is being changed
		// (example: extentions.Deployment -> apps.Deployment)
		var deleteChanges, nonDeleteChanges []Change
		for _, idx := range idxs {
			// Clear out delete change since we assume that
			// there will be 2 changes: one update, and one delete
			if changes[idx].Op() == ChangeOpDelete {
				changes[idx] = nil
				deleteChanges = append(deleteChanges, changes[idx])
			} else {
				nonDeleteChanges = append(nonDeleteChanges, changes[idx])
			}
		}

		if len(deleteChanges) != 1 {
			return nil, fmt.Errorf("Expected exactly one delete change in changes: %s",
				strings.Join(changeDescs, ", "))
		}
		if len(nonDeleteChanges) != 1 {
			return nil, fmt.Errorf("Expected exactly one non-delete change in changes: %s",
				strings.Join(changeDescs, ", "))
		}
	}

	var result []Change
	for _, change := range changes {
		if change != nil {
			result = append(result, change)
		}
	}
	return result, nil
}
