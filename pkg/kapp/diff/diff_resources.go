// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"sort"

	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

type DiffResources struct {
	ExisitingResources, NewResources versionedResources
	ExistingResourcesGrouped         map[string][]ctlres.Resource

	newRs []ctlres.Resource
	rules []ctlconf.TemplateRule
}

func NewDiffResources(existingRs, newRs []ctlres.Resource, rules []ctlconf.TemplateRule) DiffResources {
	existingVRs := existingVersionedResources(existingRs)
	newVRs := newVersionedResources(newRs)

	existingRsGrouped := newGroupedVersionedResources(existingVRs.Versioned)

	return DiffResources{existingVRs, newVRs, existingRsGrouped, newRs, rules}
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

func newGroupedVersionedResources(rs []ctlres.Resource) map[string][]ctlres.Resource {
	result := map[string][]ctlres.Resource{}

	groupByFunc := func(res ctlres.Resource) string {
		_, found := res.Annotations()[versionedResAnnKey]
		if found {
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

func (d DiffResources) existingResourcesMap() map[string]ctlres.Resource {
	result := map[string]ctlres.Resource{}

	for _, res := range d.ExisitingResources.NonVersioned {
		resKey := ctlres.NewUniqueResourceKey(res).String()
		result[resKey] = res
	}

	for resKey, res := range d.ExistingResourcesGrouped {
		result[resKey] = res[len(res)-1]
	}
	return result
}
