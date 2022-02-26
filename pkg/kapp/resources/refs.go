// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceRef struct {
	schema.GroupVersionResource
}

type PartialResourceRef struct {
	schema.GroupVersionResource
}

func (r PartialResourceRef) Matches(other schema.GroupVersionResource) bool {
	s := r.GroupVersionResource

	// TODO: support matching on Group+Resource
	// so that, for example, SpecificResFactory's can fine-tune which resources
	// are fetched.
	switch {
	case len(s.Version) > 0 && len(s.Resource) > 0:
		return s == other
	case len(s.Resource) > 0:
		return s.Group == other.Group && s.Resource == other.Resource
	case len(s.Version) > 0:
		return s.Group == other.Group && s.Version == other.Version
	case len(s.Version) == 0 && len(s.Resource) == 0:
		return s.Group == other.Group
	default:
		return false
	}
}

type GKResourceRef struct {
	schema.GroupKind
}

func (r GKResourceRef) Matches(other ResourceType) bool {
	return r.Group == other.APIResource.Group && r.Kind == other.Kind
}
