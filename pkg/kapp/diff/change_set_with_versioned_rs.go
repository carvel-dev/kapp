// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"sort"
	"strconv"

	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

const (
	versionedResAnnKey        = "kapp.k14s.io/versioned"               // Value is ignored
	versionedResOrigAnnKey    = "kapp.k14s.io/versioned-keep-original" // Value is ignored
	versionedResNumVersAnnKey = "kapp.k14s.io/num-versions"
)

type ChangeSetWithVersionedRs struct {
	existingRs, newRs         []ctlres.Resource
	rules                     []ctlconf.TemplateRule
	opts                      ChangeSetOpts
	changeFactory             ChangeFactory
	stripNameHashSuffixConfig stripNameHashSuffixConfig
}

func NewChangeSetWithVersionedRs(existingRs, newRs []ctlres.Resource,
	rules []ctlconf.TemplateRule, opts ChangeSetOpts, changeFactory ChangeFactory, stripNameHashSuffixConfigs ctlconf.StripNameHashSuffixConfigs) *ChangeSetWithVersionedRs {

	return &ChangeSetWithVersionedRs{existingRs, newRs, rules, opts, changeFactory, newStripNameHashSuffixConfigFromConf(stripNameHashSuffixConfigs)}
}

func (d ChangeSetWithVersionedRs) Calculate() ([]Change, error) {
	existingRs := d.existingVersionedResources()
	existingRsGrouped := d.groupResources(existingRs.Versioned)

	newRs := d.newVersionedResources()
	allChanges := []Change{}

	d.assignNewNames(newRs, existingRsGrouped)

	// First try to calculate changes will update references on all resources
	// (which includes versioned and non-versioned resources)
	_, _, err := d.addAndKeepChanges(newRs, existingRsGrouped)
	if err != nil {
		return nil, err
	}

	// Since there might have been circular dependencies;
	// second try catches ones that werent changed during first run
	addChanges, alreadyAdded, err := d.addAndKeepChanges(newRs, existingRsGrouped)
	if err != nil {
		return nil, err
	}

	allChanges = append(allChanges, addChanges...)

	keepAndDeleteChanges, err := d.noopAndDeleteChanges(existingRsGrouped, alreadyAdded)
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

func (d ChangeSetWithVersionedRs) groupResources(vrs []VersionedResource) map[string][]VersionedResource {
	result := map[string][]VersionedResource{}

	groupByFunc := func(ver VersionedResource) string {
		return ver.UniqVersionedKey().String()
	}

	for resKey, subVRs := range (GroupVersionedResources{vrs, groupByFunc}).Resources() {
		sort.Slice(subVRs, func(i, j int) bool {
			return subVRs[i].Version() < subVRs[j].Version()
		})
		result[resKey] = subVRs
	}

	return result
}

func (d ChangeSetWithVersionedRs) assignNewNames(
	newVRs versionedResources, existingVRsGrouped map[string][]VersionedResource) {

	// TODO name isnt used during diffing, should it?
	for _, newVerRes := range newVRs.Versioned {
		newResKey := newVerRes.UniqVersionedKey().String()

		if existingVRs, found := existingVRsGrouped[newResKey]; found {
			existingVerRes := existingVRs[len(existingVRs)-1]
			newVerRes.SetBaseName(existingVerRes.Version() + 1)
		} else {
			newVerRes.SetBaseName(1)
		}
	}
}

func (d ChangeSetWithVersionedRs) addAndKeepChanges(
	newVRs versionedResources, existingVRsGrouped map[string][]VersionedResource) (
	[]Change, map[string]VersionedResource, error) {

	changes := []Change{}
	alreadyAdded := map[string]VersionedResource{}

	for _, newVerRes := range newVRs.Versioned {
		newResKey := newVerRes.UniqVersionedKey().String()
		usedVerRes := newVerRes

		if existingVRs, found := existingVRsGrouped[newResKey]; found {
			existingVerRes := existingVRs[len(existingVRs)-1]

			// Calculate update change to determine if anything changed
			updateChange, err := d.newChange(existingVerRes.Res(), newVerRes.Res())
			if err != nil {
				return nil, nil, err
			}

			switch updateChange.Op() {
			case ChangeOpUpdate:
				changes = append(changes, d.newAddChangeFromUpdateChange(newVerRes.Res(), updateChange))
			case ChangeOpKeep:
				// Use latest copy of resource to update affected resources
				usedVerRes = existingVerRes
				changes = append(changes, d.newKeepChange(existingVerRes.Res()))
			default:
				panic(fmt.Sprintf("Unexpected change op %s", updateChange.Op()))
			}
		} else {
			// Since there no existing resource, create change for new resource
			addChange, err := d.newChange(nil, newVerRes.Res())
			if err != nil {
				return nil, nil, err
			}
			changes = append(changes, addChange)
		}

		// Update both versioned and non-versioned

		err := usedVerRes.UpdateAffected(newVRs.NonVersioned)
		if err != nil {
			return nil, nil, err
		}

		// TODO ResourceHolder?
		// err = usedVerRes.UpdateAffected(newVRs.Versioned)
		// workaround:
		newRs := []ctlres.Resource{}
		for _, newVerRes := range newVRs.Versioned {
			newRs = append(newRs, newVerRes.Res())
		}
		// workaround end
		err = usedVerRes.UpdateAffected(newRs)
		if err != nil {
			return nil, nil, err
		}

		alreadyAdded[newResKey] = newVerRes
	}

	return changes, alreadyAdded, nil
}

func (d ChangeSetWithVersionedRs) newAddChangeFromUpdateChange(
	newRes ctlres.Resource, updateChange Change) Change {

	// Use update's diffs but create a change for new resource
	return NewChangePrecalculated(nil, newRes, newRes, ChangeOpAdd, updateChange.ConfigurableTextDiff(), updateChange.OpsDiff())
}

func (d ChangeSetWithVersionedRs) noopAndDeleteChanges(
	existingVRsGrouped map[string][]VersionedResource,
	alreadyAdded map[string]VersionedResource) ([]Change, error) {

	changes := []Change{}

	// Find existing resources that were not already diffed (not in new set of resources)
	for existingResKey, existingVRs := range existingVRsGrouped {
		numToKeep := 0

		newVRes, found := alreadyAdded[existingResKey]
		if found {
			var err error
			numToKeep, err = d.numOfResourcesToKeep(newVRes)
			if err != nil {
				return nil, err
			}
		}
		if numToKeep > len(existingVRs) {
			numToKeep = len(existingVRs)
		}

		// Create changes to delete all or extra resources
		for _, existingVRes := range existingVRs[0 : len(existingVRs)-numToKeep] {

			existingRes := existingVRes.Res()

			var newRes ctlres.Resource
			if numToKeep == 0 && found && existingRes.Name() == newVRes.Res().Name() {
				// versioned resources without "last X" semantics would also
				// delete the not-actually-old existing resource when it is
				// reapplied without changes.
				newRes = newVRes.Res()
			}

			change, err := d.newChange(existingRes, newRes)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		}

		// Create changes that "noop" resources
		for _, existingVRes := range existingVRs[len(existingVRs)-numToKeep:] {
			changes = append(changes, d.newNoopChange(existingVRes.Res()))
		}
	}

	return changes, nil
}

func (d ChangeSetWithVersionedRs) newKeepChange(existingRes ctlres.Resource) Change {
	return NewChangePrecalculated(existingRes, nil, nil, ChangeOpKeep, NewConfigurableTextDiff(existingRes, nil, true), OpsDiff{})
}

func (d ChangeSetWithVersionedRs) newNoopChange(existingRes ctlres.Resource) Change {
	return NewChangePrecalculated(existingRes, nil, nil, ChangeOpNoop, nil, OpsDiff{})
}

func (ChangeSetWithVersionedRs) numOfResourcesToKeep(vres VersionedResource) (int, error) {
	switch vres.(type) {
	case HashSuffixResource:
		// there is no meaningful way to order hash-suffixed resources and as such there is no "last X" resources semantic.
		// thus simply delete all old resources.
		return 0, nil
	}

	// TODO get rid of arbitrary cut off
	numToKeep := 5
	res := vres.Res()

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
	Versioned    []VersionedResource
	NonVersioned []ctlres.Resource
}

func (d ChangeSetWithVersionedRs) newVersionedResources() versionedResources {
	var result versionedResources
	for _, res := range d.newRs {
		result.appendRes(d.VersionedFromNewResource(res))
	}
	return result
}

func (d ChangeSetWithVersionedRs) existingVersionedResources() versionedResources {
	var result versionedResources
	for _, res := range d.existingRs {
		result.appendRes(d.VersionedFromExistingResource(res))
	}
	return result
}

func (d *versionedResources) appendRes(ver VersionedResource, nonVer ctlres.Resource) {
	if ver != nil {
		d.Versioned = append(d.Versioned, ver)
	}
	if nonVer != nil {
		d.NonVersioned = append(d.NonVersioned, nonVer)
	}
}

func (d ChangeSetWithVersionedRs) VersionedFromNewResource(res ctlres.Resource) (versioned VersionedResource, nonVersioned ctlres.Resource) {

	_, hasVersionedAnn := res.Annotations()[versionedResAnnKey]
	_, hasVersionedOrigAnn := res.Annotations()[versionedResOrigAnnKey]

	if hasVersionedAnn {
		versioned = VersionedResourceImpl{res, d.rules}
		if hasVersionedOrigAnn {
			nonVersioned = res.DeepCopy()
		}
		return
	}

	if d.stripNameHashSuffixConfig.EnabledFor(res) {
		versioned = HashSuffixResource{res}
		return
	}

	return nil, res
}

func (d ChangeSetWithVersionedRs) VersionedFromExistingResource(res ctlres.Resource) (versioned VersionedResource, nonVersioned ctlres.Resource) {

	// Expect that versioned resources should not be transient
	// (Annotations may have been copied from versioned resources
	// onto transient resources for non-versioning related purposes).
	if !res.Transient() {

		_, hasVersionedAnn := res.Annotations()[versionedResAnnKey]
		if hasVersionedAnn {
			versionedRes := VersionedResourceImpl{res, d.rules}
			_, version := versionedRes.BaseNameAndVersion()
			if version != "" {
				return versionedRes, nil
			}
		}

		if d.stripNameHashSuffixConfig.EnabledFor(res) {
			return HashSuffixResource{res}, nil
		}
	}

	return nil, res
}
