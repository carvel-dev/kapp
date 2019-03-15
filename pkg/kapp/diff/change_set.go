package diff

import (
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

	return changes, nil
}
