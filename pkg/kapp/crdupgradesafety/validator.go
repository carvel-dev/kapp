// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety

import (
	"errors"
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Validation is a representation of a validation to run
// against a CRD being upgraded
type Validation interface {
	// Validate contains the actual validation logic. An error being
	// returned means validation has failed
	Validate(old, new v1.CustomResourceDefinition) error
	// Name returns a human-readable name for the validation
	Name() string
}

// ValidateFunc is a function to validate a CustomResourceDefinition
// for safe upgrades. It accepts the old and new CRDs and returns an
// error if performing an upgrade from old -> new is unsafe.
type ValidateFunc func(old, new v1.CustomResourceDefinition) error

// ValidationFunc is a helper to wrap a ValidateFunc
// as an implementation of the Validation interface
type ValidationFunc struct {
	name         string
	validateFunc ValidateFunc
}

func NewValidationFunc(name string, vfunc ValidateFunc) Validation {
	return &ValidationFunc{
		name:         name,
		validateFunc: vfunc,
	}
}

func (vf *ValidationFunc) Name() string {
	return vf.name
}

func (vf *ValidationFunc) Validate(old, new v1.CustomResourceDefinition) error {
	return vf.validateFunc(old, new)
}

type Validator struct {
	Validations []Validation
}

func (v *Validator) Validate(old, new v1.CustomResourceDefinition) error {
	validateErrs := []error{}
	for _, validation := range v.Validations {
		if err := validation.Validate(old, new); err != nil {
			formattedErr := fmt.Errorf("CustomResourceDefinition %s failed upgrade safety validation. %q validation failed: %w",
				new.Name, validation.Name(), err)

			validateErrs = append(validateErrs, formattedErr)
		}
	}
	if len(validateErrs) > 0 {
		return errors.Join(validateErrs...)
	}
	return nil
}

func NoScopeChange(old, new v1.CustomResourceDefinition) error {
	if old.Spec.Scope != new.Spec.Scope {
		return fmt.Errorf("scope changed from %q to %q", old.Spec.Scope, new.Spec.Scope)
	}
	return nil
}

func NoStoredVersionRemoved(old, new v1.CustomResourceDefinition) error {
	newVersions := sets.New[string]()
	for _, version := range new.Spec.Versions {
		if !newVersions.Has(version.Name) {
			newVersions.Insert(version.Name)
		}
	}

	for _, storedVersion := range old.Status.StoredVersions {
		if !newVersions.Has(storedVersion) {
			return fmt.Errorf("stored version %q removed", storedVersion)
		}
	}

	return nil
}
