// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type UsedGKsScope struct {
	newResources []ctlres.Resource
}

func NewUsedGKsScope(newResources []ctlres.Resource) *UsedGKsScope {
	return &UsedGKsScope{newResources}
}

func (s *UsedGKsScope) GKs() []schema.GroupKind {
	var result []schema.GroupKind

	for _, res := range s.newResources {
		result = append(result, res.GroupKind())
	}

	return result
}
