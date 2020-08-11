// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"strings"
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
	var errs []error

	uniqRs := map[string]Resource{}

	for _, res := range r.resources {
		resKey := NewUniqueResourceKey(res).String()
		if uRes, found := uniqRs[resKey]; found {
			// Check if duplicate resources are same
			if !uRes.Equal(res) {
				errs = append(errs, fmt.Errorf("Found resource '%s' multiple times with different content", res.Description()))
			}
		} else {
			uniqRs[resKey] = res
			result = append(result, res)
		}
	}

	return result, r.combinedErr(errs)
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

func (r UniqueResources) combinedErr(errs []error) error {
	if len(errs) > 0 {
		var msgs []string
		for _, err := range errs {
			msgs = append(msgs, "- "+err.Error())
		}
		return fmt.Errorf("Uniqueness errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}
