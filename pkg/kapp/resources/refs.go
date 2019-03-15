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

	switch {
	case len(s.Version) > 0 && len(s.Resource) > 0:
		return s == other
	case len(s.Version) > 0 && len(s.Resource) == 0:
		return s.Group == other.Group && s.Version == other.Version
	case len(s.Version) == 0 && len(s.Resource) == 0:
		return s.Group == other.Group
	default:
		return false
	}
}
