package resources

import (
	"fmt"
)

type UniqueResourceKey struct {
	res        Resource
	customName string
}

func NewUniqueResourceKey(res Resource) UniqueResourceKey {
	return UniqueResourceKey{res, ""}
}

func NewUniqueResourceKeyWithCustomName(res Resource, name string) UniqueResourceKey {
	return UniqueResourceKey{res, name}
}

func (k UniqueResourceKey) String() string {
	// version of the resource is not included since it will change over time
	// TODO technically resource group can be changed (true uniqueness is via UID)
	name := k.res.Name()
	if len(k.customName) > 0 {
		name = k.customName
	}
	return k.res.Namespace() + "/" + k.res.APIGroup() + "/" + k.res.Kind() + "/" + name
}

type UniqueResources struct {
	resources []Resource
}

func NewUniqueResources(resources []Resource) UniqueResources {
	return UniqueResources{resources}
}

func (r UniqueResources) Resources() ([]Resource, error) {
	var result []Resource

	uniqRs := map[string]struct{}{}
	dupRs := map[string][]Resource{}

	for _, res := range r.resources {
		resKey := NewUniqueResourceKey(res).String()
		if _, found := uniqRs[resKey]; found {
			dupRs[resKey] = append(dupRs[resKey], res)
		} else {
			uniqRs[resKey] = struct{}{}
		}
	}

	// Check if all duplicate resources are same
	for _, rs := range dupRs {
		for i := 1; i < len(rs); i++ {
			if !rs[i-1].Equal(rs[i]) {
				return nil, fmt.Errorf("Expected not to find same resource '%s' with different content", rs[i].Description())
			}
		}
	}

	// Preserve ordering and only keep one resource copy
	for _, res := range r.resources {
		resKey := NewUniqueResourceKey(res).String()
		if _, found := uniqRs[resKey]; found {
			result = append(result, res)
		}
		delete(uniqRs, resKey)
	}

	return result, nil
}

func (r UniqueResources) Match(newResources []Resource) ([]Resource, error) {
	var result []Resource
	uniqRs := map[string]struct{}{}

	for _, res := range newResources {
		uniqRs[NewUniqueResourceKey(res).String()] = struct{}{}
	}

	for _, res := range r.resources {
		resKey := NewUniqueResourceKey(res).String()
		if _, found := uniqRs[resKey]; found {
			result = append(result, res)
		}
	}

	return result, nil
}
