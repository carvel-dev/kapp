// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"sort"
	"strconv"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const (
	versionedResAnnKey        = "kapp.k14s.io/versioned"               // Value is ignored
	versionedResOrigAnnKey    = "kapp.k14s.io/versioned-keep-original" // Value is ignored
	versionedResNumVersAnnKey = "kapp.k14s.io/num-versions"
)

type ChangeSetWithVersionedRs struct {
	existingRs, newRs []ctlres.Resource
	rules             []ctlconf.TemplateRule
	opts              ChangeSetOpts
	changeFactory     ChangeFactory
}

func NewChangeSetWithVersionedRs(existingRs, newRs []ctlres.Resource,
	rules []ctlconf.TemplateRule, opts ChangeSetOpts, changeFactory ChangeFactory) *ChangeSetWithVersionedRs {

	return &ChangeSetWithVersionedRs{existingRs, newRs, rules, opts, changeFactory}
}

func (d ChangeSetWithVersionedRs) Calculate() ([]Change, error) {
	existingRs := existingVersionedResources(d.existingRs)
	existingRsGrouped := d.groupResources(existingRs.Versioned)

	newRs := newVersionedResources(d.newRs)
	allChanges := []Change{}

	d.assignNewNames(newRs, existingRsGrouped)

	// First try to calculate changes will update references on all resources
	// (which includes versioned and non-versioned resources)
	_, _, err := d.addChanges(newRs, existingRsGrouped)
	if err != nil {
		return nil, err
	}

	// Since there might have been circular dependencies;
	// second try catches ones that werent changed during first run
	addChanges, alreadyAdded, err := d.addChanges(newRs, existingRsGrouped)
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, addChanges...)

	keepAndDeleteChanges, err := d.keepAndDeleteChanges(existingRsGrouped, alreadyAdded)
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, keepAndDeleteChanges...)

	nonVersionedChangeSet := NewChangeSet(
		existingRs.NonVersioned, newRs.NonVersioned, d.opts, d.changeFactory)

	nonVersionedChanges, err := nonVersionedChangeSet.Calculate()
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, nonVersionedChanges...)

	return allChanges, nil
}

func (d ChangeSetWithVersionedRs) groupResources(rs []ctlres.Resource) map[string][]ctlres.Resource {
	result := map[string][]ctlres.Resource{}

	groupByFunc := func(res ctlres.Resource) string {
		if _, found := res.Annotations()[versionedResAnnKey]; found {
			return VersionedResource{res, nil}.UniqVersionedKey().String()
		}
		panic("Expected to find versioned annotation on resource")
	}

	for resKey, subRs := range (GroupResources{rs, groupByFunc}).Resources() {
		sort.Slice(subRs, func(i, j int) bool {
			return VersionedResource{subRs[i], nil}.Version() < VersionedResource{subRs[j], nil}.Version()
		})
		result[resKey] = subRs
	}

	return result
}

func (d ChangeSetWithVersionedRs) assignNewNames(
	newRs versionedResources, existingRsGrouped map[string][]ctlres.Resource) {

	// TODO name isnt used during diffing, should it?
	for _, newRes := range newRs.Versioned {
		newVerRes := VersionedResource{newRes, nil}
		newResKey := newVerRes.UniqVersionedKey().String()

		if existingRs, found := existingRsGrouped[newResKey]; found {
			existingRes := existingRs[len(existingRs)-1]
			newVerRes.SetBaseName(VersionedResource{existingRes, nil}.Version() + 1)
		} else {
			newVerRes.SetBaseName(1)
		}
	}
}

func (d ChangeSetWithVersionedRs) addChanges(
	newRs versionedResources, existingRsGrouped map[string][]ctlres.Resource) (
	[]Change, map[string]ctlres.Resource, error) {

	changes := []Change{}
	alreadyAdded := map[string]ctlres.Resource{}

	for _, newRes := range newRs.Versioned {
		newResKey := VersionedResource{newRes, nil}.UniqVersionedKey().String()
		usedRes := newRes

		if existingRs, found := existingRsGrouped[newResKey]; found {
			existingRes := existingRs[len(existingRs)-1]

			// Calculate update change to determine if anything changed
			updateChange, err := d.newChange(existingRes, newRes)
			if err != nil {
				return nil, nil, err
			}

			switch updateChange.Op() {
			case ChangeOpUpdate:
				changes = append(changes, d.newAddChangeFromUpdateChange(newRes, updateChange))
			case ChangeOpKeep:
				// Use latest copy of resource to update affected resources
				usedRes = existingRes
			default:
				panic(fmt.Sprintf("Unexpected change op %s", updateChange.Op()))
			}
		} else {
			// Since there no existing resource, create change for new resource
			addChange, err := d.newChange(nil, newRes)
			if err != nil {
				return nil, nil, err
			}
			changes = append(changes, addChange)
		}

		// Update both versioned and non-versioned
		verRes := VersionedResource{usedRes, d.rules}

		err := verRes.UpdateAffected(newRs.NonVersioned)
		if err != nil {
			return nil, nil, err
		}

		err = verRes.UpdateAffected(newRs.Versioned)
		if err != nil {
			return nil, nil, err
		}

		alreadyAdded[newResKey] = newRes
	}

	return changes, alreadyAdded, nil
}

func (d ChangeSetWithVersionedRs) newAddChangeFromUpdateChange(
	newRes ctlres.Resource, updateChange Change) Change {

	// Use update's diffs but create a change for new resource
	addChange := NewChangePrecalculated(nil, newRes, newRes)
	// TODO private field access
	addChange.op = ChangeOpAdd
	addChange.configurableTextDiff = updateChange.ConfigurableTextDiff()
	addChange.opsDiff = updateChange.OpsDiff()
	return addChange
}

func (d ChangeSetWithVersionedRs) keepAndDeleteChanges(
	existingRsGrouped map[string][]ctlres.Resource,
	alreadyAdded map[string]ctlres.Resource) ([]Change, error) {

	changes := []Change{}

	// Find existing resources that were not already diffed (not in new set of resources)
	for existingResKey, existingRs := range existingRsGrouped {
		numToKeep := 0

		if newRes, found := alreadyAdded[existingResKey]; found {
			var err error
			numToKeep, err = d.numOfResourcesToKeep(newRes)
			if err != nil {
				return nil, err
			}
		}
		if numToKeep > len(existingRs) {
			numToKeep = len(existingRs)
		}

		// Create changes to delete all or extra resources
		for _, existingRes := range existingRs[0 : len(existingRs)-numToKeep] {
			change, err := d.newChange(existingRes, nil)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		}

		// Create changes that "keep" resources
		for _, existingRes := range existingRs[len(existingRs)-numToKeep:] {
			changes = append(changes, d.newKeepChange(existingRes))
		}
	}

	return changes, nil
}

func (d ChangeSetWithVersionedRs) newKeepChange(existingRes ctlres.Resource) Change {
	// Use update's diffs but create a change for new resource
	addChange := NewChangePrecalculated(existingRes, nil, nil)
	// TODO private field access
	addChange.op = ChangeOpKeep
	return addChange
}

func (ChangeSetWithVersionedRs) numOfResourcesToKeep(res ctlres.Resource) (int, error) {
	// TODO get rid of arbitrary cut off
	numToKeep := 5

	if numToKeepAnn, found := res.Annotations()[versionedResNumVersAnnKey]; found {
		var err error
		numToKeep, err = strconv.Atoi(numToKeepAnn)
		if err != nil {
			return 0, fmt.Errorf("Expected annotation '%s' value to be an integer", versionedResNumVersAnnKey)
		}
		if numToKeep < 1 {
			return 0, fmt.Errorf("Expected annotation '%s' value to be a >= 1", versionedResNumVersAnnKey)
		}
	} else {
		numToKeep = 5
	}

	return numToKeep, nil
}

func (d ChangeSetWithVersionedRs) newChange(existingRes, newRes ctlres.Resource) (Change, error) {
	changeFactoryFunc := d.changeFactory.NewExactChange
	if d.opts.AgainstLastApplied {
		changeFactoryFunc = d.changeFactory.NewChangeAgainstLastApplied
	}
	return changeFactoryFunc(existingRes, newRes)
}

type versionedResources struct {
	Versioned    []ctlres.Resource
	NonVersioned []ctlres.Resource
}

func newVersionedResources(rs []ctlres.Resource) versionedResources {
	var result versionedResources
	for _, res := range rs {
		_, hasVersionedAnn := res.Annotations()[versionedResAnnKey]
		_, hasVersionedOrigAnn := res.Annotations()[versionedResOrigAnnKey]

		if hasVersionedAnn {
			result.Versioned = append(result.Versioned, res)
			if hasVersionedOrigAnn {
				result.NonVersioned = append(result.NonVersioned, res.DeepCopy())
			}
		} else {
			result.NonVersioned = append(result.NonVersioned, res)
		}
	}
	return result
}

func existingVersionedResources(rs []ctlres.Resource) versionedResources {
	var result versionedResources
	for _, res := range rs {
		// Expect that versioned resources should not be transient
		// (Annotations may have been copied from versioned resources
		// onto transient resources for non-versioning related purposes).
		_, hasVersionedAnn := res.Annotations()[versionedResAnnKey]

		versionedRs := VersionedResource{res: res}
		_, version := versionedRs.BaseNameAndVersion()

		if hasVersionedAnn && !res.Transient() && version != "" {
			result.Versioned = append(result.Versioned, res)
		} else {
			result.NonVersioned = append(result.NonVersioned, res)
		}
	}
	return result
}
