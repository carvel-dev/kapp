// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
)

type VersionedRefDesc struct {
	Namespace string `json:"namespace"`
	APIGroup  string `json:"apiGroup"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
}

type ExplicitVersionedRefAnn struct {
	References     []VersionedRefDesc `json:"references"`
	VersionedNames map[string]string
}

type ExplicitVersionedRef struct {
	Resource   VersionedResource
	Annotation ExplicitVersionedRefAnn
}

const (
	explicitReferenceKey = "kapp.k14s.io/versioned-explicit-ref"
)

func NewExplicitVersionedRef(res VersionedResource, annotation ExplicitVersionedRefAnn) *ExplicitVersionedRef {
	return &ExplicitVersionedRef{res, annotation}
}

// Returns true if the resource is referenced by the annotation
func (e *ExplicitVersionedRef) IsReferenced() (bool, error) {
	references, err := e.references()
	if err != nil {
		return false, err
	}

	referenceKey := e.referenceKey()

	for _, v := range references {
		if v == referenceKey {
			return true, nil
		}
	}

	return false, nil
}

func (e *ExplicitVersionedRef) references() ([]string, error) {
	list := []string{}
	for _, v := range e.Annotation.References {
		v, err := e.validateAndReplaceCoreGroup(v)
		if err != nil {
			return list, err
		}

		list = append(list, fmt.Sprintf("%s/%s/%s/%s", v.Namespace, v.APIGroup, v.Kind, v.Name))
	}
	return list, nil
}

func (e *ExplicitVersionedRef) validateAndReplaceCoreGroup(resourceDescription VersionedRefDesc) (VersionedRefDesc, error) {
	// Replacing APIGroup value "core" with an empty string
	// Edgecase for resources that are a part of "core"
	if resourceDescription.APIGroup == "core" {
		resourceDescription.APIGroup = ""
	}

	if resourceDescription.Kind == "" || resourceDescription.Name == "" {
		return resourceDescription, fmt.Errorf("Explicit Reference Error: Name and Kind are required values in an explicit reference")
	}
	return resourceDescription, nil
}

func (e *ExplicitVersionedRef) referenceKey() string {
	return e.Resource.UniqVersionedKey().String()
}

/*
Annotation with a list of references in JSON format:

kapp.k14s.io/versioned-explicit-ref: '{ "references": [ { "namespace": <resource-namespace>, "apiGroup":  <resource-apiGroup>, "kind" : <resource-kind>, "name":  <resource-name> } ] }'

"namespace" need not be assigned a value for cluster-scoped resources.
"apiGroup" need not be assigned a value for resources belonging to "core" API group.
*/
