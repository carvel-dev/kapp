// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

type VersionedRefDesc struct {
	Namespace  string `yaml:"namespace"`
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
}

type ExplicitVersionedRef struct {
	Resource   VersionedResource
	Annotation string
}

const (
	explicitReferenceKey       = "kapp.k14s.io/versioned-explicit-ref"
	explicitReferenceKeyPrefix = "kapp.k14s.io/versioned-explicit-ref."
)

func NewExplicitVersionedRef(res VersionedResource, annotation string) *ExplicitVersionedRef {
	return &ExplicitVersionedRef{res, annotation}
}

// Returns true if the resource is referenced by the annotation
func (e *ExplicitVersionedRef) IsReferenced() (bool, error) {
	reference := VersionedRefDesc{}
	err := yaml.Unmarshal([]byte(e.Annotation), &reference)
	if err != nil {
		return false, fmt.Errorf("Error unmarshalling versioned reference: %s", err)
	}

	if reference.APIVersion == "" || reference.Kind == "" || reference.Name == "" {
		return false, fmt.Errorf("Explicit reference error: apiVersion, kind and name are required values in an explicit versioned reference")
	}

	baseName, _ := e.Resource.BaseNameAndVersion()
	versionedResourceDesc := VersionedRefDesc{
		Namespace:  e.Resource.res.Namespace(),
		APIVersion: e.Resource.res.APIVersion(),
		Kind:       e.Resource.res.Kind(),
		Name:       baseName,
	}

	return reflect.DeepEqual(versionedResourceDesc, reference), nil
}

func (e *ExplicitVersionedRef) VersionedReference() (string, error) {
	reference := VersionedRefDesc{}
	err := yaml.Unmarshal([]byte(e.Annotation), &reference)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling versioned reference: %s", err)
	}

	reference.Name = e.Resource.res.Name()
	versionedReference, err := yaml.Marshal(reference)
	if err != nil {
		return "", fmt.Errorf("Error marshalling versioned reference: %s", err)
	}

	return string(versionedReference), nil
}

/*
Annotation with a list of references in JSON format:

kapp.k14s.io/versioned-explicit-ref: '{ "references": [ { "namespace": <resource-namespace>, "apiGroup":  <resource-apiGroup>, "kind" : <resource-kind>, "name":  <resource-name> } ] }'

"namespace" need not be assigned a value for cluster-scoped resources.
"apiGroup" need not be assigned a value for resources belonging to "core" API group.
*/
