// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety

import (
	"errors"
	"fmt"
	"strings"

	"github.com/openshift/crd-schema-checker/pkg/manifestcomparators"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Validation is a representation of a validation to run
// against a CRD being upgraded
type Validation interface {
	// Validate contains the actual validation logic. An error being
	// returned means validation has failed
	Validate(oldCRD, newCRD v1.CustomResourceDefinition) error
	// Name returns a human-readable name for the validation
	Name() string
}

// ValidateFunc is a function to validate a CustomResourceDefinition
// for safe upgrades. It accepts the old and new CRDs and returns an
// error if performing an upgrade from old -> new is unsafe.
type ValidateFunc func(oldCRD, newCRD v1.CustomResourceDefinition) error

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

func (vf *ValidationFunc) Validate(oldCRD, newCRD v1.CustomResourceDefinition) error {
	return vf.validateFunc(oldCRD, newCRD)
}

type Validator struct {
	Validations []Validation
}

func (v *Validator) Validate(oldCRD, newCRD v1.CustomResourceDefinition) error {
	validateErrs := []error{}
	for _, validation := range v.Validations {
		if err := validation.Validate(oldCRD, newCRD); err != nil {
			formattedErr := fmt.Errorf("CustomResourceDefinition %s failed upgrade safety validation. %q validation failed: %w",
				newCRD.Name, validation.Name(), err)

			validateErrs = append(validateErrs, formattedErr)
		}
	}
	if len(validateErrs) > 0 {
		return errors.Join(validateErrs...)
	}
	return nil
}

func NoScopeChange(oldCRD, newCRD v1.CustomResourceDefinition) error {
	if oldCRD.Spec.Scope != newCRD.Spec.Scope {
		return fmt.Errorf("scope changed from %q to %q", oldCRD.Spec.Scope, newCRD.Spec.Scope)
	}
	return nil
}

func NoStoredVersionRemoved(oldCRD, newCRD v1.CustomResourceDefinition) error {
	newVersions := sets.New[string]()
	for _, version := range newCRD.Spec.Versions {
		if !newVersions.Has(version.Name) {
			newVersions.Insert(version.Name)
		}
	}

	for _, storedVersion := range oldCRD.Status.StoredVersions {
		if !newVersions.Has(storedVersion) {
			return fmt.Errorf("stored version %q removed", storedVersion)
		}
	}

	return nil
}

func NoExistingFieldRemoved(oldCRD, newCRD v1.CustomResourceDefinition) error {
	reg := manifestcomparators.NewRegistry()
	err := reg.AddComparator(manifestcomparators.NoFieldRemoval())
	if err != nil {
		return err
	}

	results, errs := reg.Compare(&oldCRD, &newCRD)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	errSet := []error{}

	for _, result := range results {
		if len(result.Errors) > 0 {
			errSet = append(errSet, errors.New(strings.Join(result.Errors, "\n")))
		}
	}
	if len(errSet) > 0 {
		return errors.Join(errSet...)
	}

	return nil
}
