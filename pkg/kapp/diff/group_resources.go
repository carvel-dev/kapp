// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type GroupResources struct {
	resources []ctlres.Resource
	groupFunc func(ctlres.Resource) string
}

func (r GroupResources) Resources() map[string][]ctlres.Resource {
	result := map[string][]ctlres.Resource{}

	for _, res := range r.resources {
		id := r.groupFunc(res)
		result[id] = append(result[id], res)
	}

	return result
}
