// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"gopkg.in/yaml.v2"
)

type ExplicitVersionedRef struct {
	AnnotationKey string
	Annotation    string
}

const (
	explicitReferenceKey       = "kapp.k14s.io/versioned-explicit-ref"
	explicitReferenceKeyPrefix = "kapp.k14s.io/versioned-explicit-ref."
)

func NewExplicitVersionedRef(annotationKey string, annotation string) *ExplicitVersionedRef {
	return &ExplicitVersionedRef{annotationKey, annotation}
}

func (e *ExplicitVersionedRef) AsObjectRef() (map[string]interface{}, error) {
	var objectRef map[string]interface{}
	err := yaml.Unmarshal([]byte(e.Annotation), &objectRef)
	if err != nil {
		return nil, fmt.Errorf("Parsing versioned explicit reference from annotation '%s': %w", e.AnnotationKey, err)
	}

	_, hasAPIVersionKey := objectRef["apiVersion"]
	_, hasKindKey := objectRef["kind"]
	_, hasNameKey := objectRef["name"]

	if !(hasAPIVersionKey && hasKindKey && hasNameKey) {
		return nil, fmt.Errorf("Expected versioned explicit reference to specify non-empty apiVersion, kind and name keys")
	}

	return objectRef, nil
}

func (e *ExplicitVersionedRef) AnnotationMod(objectRef map[string]interface{}) (ctlres.StringMapAppendMod, error) {
	value, err := yaml.Marshal(objectRef)
	if err != nil {
		return ctlres.StringMapAppendMod{}, fmt.Errorf("Marshalling explicit reference: %w", err)
	}

	return ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs: map[string]string{
			e.AnnotationKey: string(value),
		},
	}, nil
}
