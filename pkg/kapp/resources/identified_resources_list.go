// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r IdentifiedResources) List(labelSelector labels.Selector, resRefs []ResourceRef) ([]Resource, error) {
	defer r.logger.DebugFunc("List").Finish()

	resTypes, err := r.resourceTypes.All()
	if err != nil {
		return nil, err
	}

	// TODO non-listable types
	resTypes = Listable(resTypes)

	// TODO eliminating events
	resTypes = NonMatching(resTypes, ResourceRef{
		schema.GroupVersionResource{Version: "v1", Resource: "events"},
	})

	// TODO eliminating component statuses
	resTypes = NonMatching(resTypes, ResourceRef{
		schema.GroupVersionResource{Version: "v1", Resource: "componentstatuses"},
	})

	if len(resRefs) > 0 {
		resTypes = MatchingAny(resTypes, resRefs)
	}

	allOpts := ResourcesAllOpts{
		ListOpts: &metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	}

	resources, err := r.resources.All(resTypes, allOpts)
	if err != nil {
		return nil, err
	}

	// Mark resources that were not created by kapp as transient
	for i, res := range resources {
		if !NewIdentityAnnotation(res).Valid() {
			res.MarkTransient(true)
			resources[i] = res
		}
	}

	return r.pickPreferredVersions(resources)
}

func (r IdentifiedResources) pickPreferredVersions(resources []Resource) ([]Resource, error) {
	var result []Resource

	uniqueByID := map[string][]Resource{}

	for _, res := range resources {
		uniqueByID[res.UID()] = append(uniqueByID[res.UID()], res)
	}

	for _, rs := range uniqueByID {
		var matched bool

		for _, res := range rs {
			idAnn := NewIdentityAnnotation(res)

			if idAnn.MatchesVersion() {
				err := idAnn.RemoveMod().Apply(res)
				if err != nil {
					return nil, err
				}

				result = append(result, res)
				matched = true
				break
			}
		}

		if !matched {
			// Sort to have some stability
			sort.Slice(rs, func(i, j int) bool { return rs[i].APIVersion() < rs[j].APIVersion() })
			// TODO use preferred version from the api
			result = append(result, rs[0])
		}
	}

	return result, nil
}
