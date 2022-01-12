// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type UsedGVKsScope struct {
	newResources []ctlres.Resource
}

func NewUsedGVsScope(newResources []ctlres.Resource) *UsedGVKsScope {
	return &UsedGVKsScope{newResources}
}

func (s *UsedGVKsScope) GVKs() []schema.GroupVersionKind {
	var result []schema.GroupVersionKind

	for _, res := range s.newResources {
		result = append(result, res.GroupVersionKind())
	}

	return result
}
