// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

type GroupVersionedResources struct {
	resources []VersionedResource
	groupFunc func(VersionedResource) string
}

func (r GroupVersionedResources) Resources() map[string][]VersionedResource {
	result := map[string][]VersionedResource{}

	for _, res := range r.resources {
		id := r.groupFunc(res)
		result[id] = append(result[id], res)
	}

	return result
}
