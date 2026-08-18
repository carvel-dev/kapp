// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/openshift/crd-schema-checker/pkg/manifestcomparators"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ChangeValidation is a function that accepts a FieldDiff
// as a parameter and should return:
// - a boolean representation of whether or not the change
// - an error if the change would be unsafe
// has been fully handled (i.e no additional changes exist)
type ChangeValidation func(diff FieldDiff) (bool, error)

// EnumChangeValidation ensures that:
// - No enums are added to a field that did not previously have
// enum restrictions
// - No enums are removed from a field
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e the only change was to enum values)
// - An error if either of the above validations are not satisfied
func EnumChangeValidation(diff FieldDiff) (bool, error) {
	// This function resets the enum values for the
	// old and new field and compares them to determine
	// if there are any additional changes that should be
	// handled. Reseting the enum values allows for chained
	// evaluations to check if they have handled all the changes
	// without having to account for fields other than the ones
	// they are designed to handle. This function should only be called when
	// returning from this function to prevent unnecessary overwrites of
	// these fields.
	handled := func() bool {
		diff.Old.Enum = []v1.JSON{}
		diff.New.Enum = []v1.JSON{}
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	if len(diff.Old.Enum) == 0 && len(diff.New.Enum) > 0 {
		return handled(), fmt.Errorf("enums added when there were no enum restrictions previously")
	}

	oldSet := sets.NewString()
	for _, enum := range diff.Old.Enum {
		if !oldSet.Has(string(enum.Raw)) {
			oldSet.Insert(string(enum.Raw))
		}
	}

	newSet := sets.NewString()
	for _, enum := range diff.New.Enum {
		if !newSet.Has(string(enum.Raw)) {
			newSet.Insert(string(enum.Raw))
		}
	}

	diffSet := oldSet.Difference(newSet)
	if diffSet.Len() > 0 {
		return handled(), fmt.Errorf("enum values removed: %+v", diffSet.UnsortedList())
	}

	return handled(), nil
}

// RequiredFieldChangeValidation adds a validation check to ensure that
// existing required fields can be marked as optional in a CRD schema:
// - No new values can be added as required that did not previously have
// any required fields present
// - Existing values can be removed from the required field
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to required field values)
// - An error if either of the above criteria are not met
func RequiredFieldChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.Required = []string{}
		diff.New.Required = []string{}
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	if len(diff.Old.Required) == 0 && len(diff.New.Required) > 0 {
		return handled(), fmt.Errorf("new values added as required when previously no required fields existed: %+v", diff.New.Required)
	}

	oldSet := sets.NewString()
	for _, requiredField := range diff.Old.Required {
		if !oldSet.Has(requiredField) {
			oldSet.Insert(requiredField)
		}
	}

	newSet := sets.NewString()
	for _, requiredField := range diff.New.Required {
		if !newSet.Has(requiredField) {
			newSet.Insert(requiredField)
		}
	}

	diffSet := newSet.Difference(oldSet)
	if diffSet.Len() > 0 {
		return handled(), fmt.Errorf("new required fields added: %+v", diffSet.UnsortedList())
	}

	return handled(), nil
}

// MinimumChangeValidation adds a validation check to ensure that
// existing fields can have their minimum constraints updated in a CRD schema
// based on the following:
// - No minimum constraint can be added if one did not exist previously
// - Minimum constraints can not increase in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to minimum constraints)
// - An error if either of the above criteria are not met
func MinimumChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.Minimum = nil
		diff.New.Minimum = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.Minimum == nil && diff.New.Minimum != nil:
		m := *diff.New.Minimum
		return handled(), fmt.Errorf("minimum constraint added when one did not exist previously: %+v", m)
	case diff.Old.Minimum != nil && diff.New.Minimum != nil:
		oldMin := *diff.Old.Minimum
		newMin := *diff.New.Minimum
		if oldMin < newMin {
			return handled(), fmt.Errorf("minimum constraint increased from %+v to %+v", oldMin, newMin)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MinimumLengthChangeValidation adds a validation check to ensure that
// existing fields can have their minimum length constraints updated in a CRD schema
// based on the following:
// - No minimum length constraint can be added if one did not exist previously
// - Minimum length constraints can not increase in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to minimum length constraints)
// - An error if either of the above criteria are not met
func MinimumLengthChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.MinLength = nil
		diff.New.MinLength = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.MinLength == nil && diff.New.MinLength != nil:
		m := *diff.New.MinLength
		return handled(), fmt.Errorf("minimum length constraint added when one did not exist previously: %+v", m)
	case diff.Old.MinLength != nil && diff.New.MinLength != nil:
		oldMin := *diff.Old.MinLength
		newMin := *diff.New.MinLength
		if oldMin < newMin {
			return handled(), fmt.Errorf("minimum length constraint increased from %+v to %+v", oldMin, newMin)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MinimumItemsChangeValidation adds a validation check to ensure that
// existing fields can have their minimum item constraints updated in a CRD schema
// based on the following:
// - No minimum item constraint can be added if one did not exist previously
// - Minimum item constraints can not increase in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to minimum item constraints)
// - An error if either of the above criteria are not met
func MinimumItemsChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.MinItems = nil
		diff.New.MinItems = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.MinItems == nil && diff.New.MinItems != nil:
		m := *diff.New.MinItems
		return handled(), fmt.Errorf("minimum items constraint added when one did not exist previously: %+v", m)
	case diff.Old.MinItems != nil && diff.New.MinItems != nil:
		oldMin := *diff.Old.MinItems
		newMin := *diff.New.MinItems
		if oldMin < newMin {
			return handled(), fmt.Errorf("minimum items constraint increased from %+v to %+v", oldMin, newMin)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MinimumPropertiesChangeValidation adds a validation check to ensure that
// existing fields can have their minimum properties constraints updated in a CRD schema
// based on the following:
// - No minimum properties constraint can be added if one did not exist previously
// - Minimum properties constraints can not increase in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to minimum properties constraints)
// - An error if either of the above criteria are not met
func MinimumPropertiesChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.MinProperties = nil
		diff.New.MinProperties = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.MinProperties == nil && diff.New.MinProperties != nil:
		m := *diff.New.MinProperties
		return handled(), fmt.Errorf("minimum properties constraint added when one did not exist previously: %+v", m)
	case diff.Old.MinProperties != nil && diff.New.MinProperties != nil:
		oldMin := *diff.Old.MinProperties
		newMin := *diff.New.MinProperties
		if oldMin < newMin {
			return handled(), fmt.Errorf("minimum properties constraint increased from %+v to %+v", oldMin, newMin)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MaximumChangeValidation adds a validation check to ensure that
// existing fields can have their maximum constraints updated in a CRD schema
// based on the following:
// - No maximum constraint can be added if one did not exist previously
// - Maximum constraints can not decrease in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to maximum constraints)
// - An error if either of the above criteria are not met
func MaximumChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.Maximum = nil
		diff.New.Maximum = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.Maximum == nil && diff.New.Maximum != nil:
		m := *diff.New.Maximum
		return handled(), fmt.Errorf("maximum constraint added when one did not exist previously: %+v", m)
	case diff.Old.Maximum != nil && diff.New.Maximum != nil:
		oldMax := *diff.Old.Maximum
		newMax := *diff.New.Maximum
		if newMax < oldMax {
			return handled(), fmt.Errorf("maximum constraint decreased from %+v to %+v", oldMax, newMax)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MaximumLengthChangeValidation adds a validation check to ensure that
// existing fields can have their maximum length constraints updated in a CRD schema
// based on the following:
// - No maximum length constraint can be added if one did not exist previously
// - Maximum length constraints can not decrease in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to maximum length constraints)
// - An error if either of the above criteria are not met
func MaximumLengthChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.MaxLength = nil
		diff.New.MaxLength = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.MaxLength == nil && diff.New.MaxLength != nil:
		m := *diff.New.MaxLength
		return handled(), fmt.Errorf("maximum length constraint added when one did not exist previously: %+v", m)
	case diff.Old.MaxLength != nil && diff.New.MaxLength != nil:
		oldMax := *diff.Old.MaxLength
		newMax := *diff.New.MaxLength
		if newMax < oldMax {
			return handled(), fmt.Errorf("maximum length constraint decreased from %+v to %+v", oldMax, newMax)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MaximumItemsChangeValidation adds a validation check to ensure that
// existing fields can have their maximum item constraints updated in a CRD schema
// based on the following:
// - No maximum item constraint can be added if one did not exist previously
// - Maximum item constraints can not decrease in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to maximum item constraints)
// - An error if either of the above criteria are not met
func MaximumItemsChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.MaxItems = nil
		diff.New.MaxItems = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.MaxItems == nil && diff.New.MaxItems != nil:
		m := *diff.New.MaxItems
		return handled(), fmt.Errorf("maximum items constraint added when one did not exist previously: %+v", m)
	case diff.Old.MaxItems != nil && diff.New.MaxItems != nil:
		oldMax := *diff.Old.MaxItems
		newMax := *diff.New.MaxItems
		if newMax < oldMax {
			return handled(), fmt.Errorf("maximum items constraint decreased from %+v to %+v", oldMax, newMax)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// MaximumPropertiesChangeValidation adds a validation check to ensure that
// existing fields can have their maximum properties constraints updated in a CRD schema
// based on the following:
// - No maximum properties constraint can be added if one did not exist previously
// - Maximum properties constraints can not increase in value
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to maximum properties constraints)
// - An error if either of the above criteria are not met
func MaximumPropertiesChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.MaxProperties = nil
		diff.New.MaxProperties = nil
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.MaxProperties == nil && diff.New.MaxProperties != nil:
		m := *diff.New.MaxProperties
		return handled(), fmt.Errorf("maximum properties constraint added when one did not exist previously: %+v", m)
	case diff.Old.MaxProperties != nil && diff.New.MaxProperties != nil:
		oldMax := *diff.Old.MaxProperties
		newMax := *diff.New.MaxProperties
		if newMax < oldMax {
			return handled(), fmt.Errorf("maximum properties constraint decreased from %+v to %+v", oldMax, newMax)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// DefaultValueChangeValidation adds a validation check to ensure that
// default values are not changed in a CRD schema:
// - No new value can be added as default that did not previously have a
// default value present
// - Default value of a field cannot be changed
// - Existing default value for a field cannot be removed
// This function returns:
// - A boolean representation of whether or not the change
// has been fully handled (i.e. the only change was to a field's default value)
// - An error if either of the above criteria are not met
func DefaultValueChangeValidation(diff FieldDiff) (bool, error) {
	handled := func() bool {
		diff.Old.Default = &v1.JSON{}
		diff.New.Default = &v1.JSON{}
		return reflect.DeepEqual(diff.Old, diff.New)
	}

	switch {
	case diff.Old.Default == nil && diff.New.Default != nil:
		newDefault := diff.New.Default
		return handled(), fmt.Errorf("new value added as default when previously no default value existed: %+v", newDefault)

	case diff.Old.Default != nil && diff.New.Default == nil:
		oldDefault := diff.Old.Default.Raw
		return handled(), fmt.Errorf("default value has been removed when previously a default value existed: %+v", oldDefault)

	case diff.Old.Default != nil && diff.New.Default != nil:
		oldDefault := diff.Old.Default.Raw
		newDefault := diff.New.Default.Raw
		if !bytes.Equal(diff.Old.Default.Raw, diff.New.Default.Raw) {
			return handled(), fmt.Errorf("default value has been changed from %+v to %+v", oldDefault, newDefault)
		}
		fallthrough
	default:
		return handled(), nil
	}
}

// ChangeValidator is a Validation implementation focused on
// handling updates to existing fields in a CRD
type ChangeValidator struct {
	// Validations is a slice of ChangeValidations
	// to run against each changed field
	Validations []ChangeValidation
}

func (cv *ChangeValidator) Name() string {
	return "ChangeValidator"
}

// Validate will compare each version in the provided existing and new CRDs.
// Since the ChangeValidator is tailored to handling updates to existing fields in
// each version of a CRD. As such the following is assumed:
// - Validating the removal of versions during an update is handled outside of this
// validator. If a version in the existing version of the CRD does not exist in the new
// version that version of the CRD is skipped in this validator.
// - Removal of existing fields is unsafe. Regardless of whether or not this is handled
// by a validator outside this one, if a field is present in a version provided by the existing CRD
// but not present in the same version provided by the new CRD this validation will fail.
//
// Additionally, any changes that are not validated and handled by the known ChangeValidations
// are deemed as unsafe and returns an error.
func (cv *ChangeValidator) Validate(oldCRD, newCRD v1.CustomResourceDefinition) error {
	errs := []error{}
	for _, version := range oldCRD.Spec.Versions {
		newVersion := manifestcomparators.GetVersionByName(&newCRD, version.Name)
		if newVersion == nil {
			// if the new version doesn't exist skip this version
			continue
		}
		flatOld := FlattenSchema(version.Schema.OpenAPIV3Schema)
		flatNew := FlattenSchema(newVersion.Schema.OpenAPIV3Schema)

		diffs, err := CalculateFlatSchemaDiff(flatOld, flatNew)
		if err != nil {
			errs = append(errs, fmt.Errorf("calculating schema diff for CRD version %q", version.Name))
			continue
		}

		// ensure order of the potentially multi-line final error
		for _, field := range slices.Sorted(maps.Keys(diffs)) {
			diff := diffs[field]

			handled := false
			for _, validation := range cv.Validations {
				ok, err := validation(diff)
				if err != nil {
					errs = append(errs, fmt.Errorf("version %q, field %q: %w", version.Name, field, err))
				}
				if ok {
					handled = true
					break
				}
			}

			if !handled {
				errs = append(errs, fmt.Errorf("version %q, field %q has unknown change, refusing to determine that change is safe", version.Name, field))
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

type FieldDiff struct {
	Old *v1.JSONSchemaProps
	New *v1.JSONSchemaProps
}

// FlatSchema is a flat representation of a CRD schema.
type FlatSchema map[string]*v1.JSONSchemaProps

// FlattenSchema takes in a CRD version OpenAPIV3Schema and returns
// a flattened representation of it. For example, a CRD with a schema of:
// ```yaml
//
//	...
//	spec:
//	  type: object
//	  properties:
//	    foo:
//	      type: string
//	    bar:
//	      type: string
//	...
//
// ```
// would be represented as:
//
//	map[string]*v1.JSONSchemaProps{
//	   "^": {},
//	   "^.spec": {},
//	   "^.spec.foo": {},
//	   "^.spec.bar": {},
//	}
//
// where "^" represents the "root" schema
func FlattenSchema(schema *v1.JSONSchemaProps) FlatSchema {
	fieldMap := map[string]*v1.JSONSchemaProps{}

	manifestcomparators.SchemaHas(schema,
		field.NewPath("^"),
		field.NewPath("^"),
		nil,
		func(s *v1.JSONSchemaProps, _, simpleLocation *field.Path, _ []*v1.JSONSchemaProps) bool {
			fieldMap[simpleLocation.String()] = s.DeepCopy()
			return false
		})

	return fieldMap
}

// CalculateFlatSchemaDiff finds fields in a FlatSchema that are different
// and returns a mapping of field --> old and new field schemas. If a field
// exists in the old FlatSchema but not the new an empty diff mapping and an error is returned.
func CalculateFlatSchemaDiff(o, n FlatSchema) (map[string]FieldDiff, error) {
	diffMap := map[string]FieldDiff{}
	for field, schema := range o {
		if _, ok := n[field]; !ok {
			return diffMap, fmt.Errorf("field %q in existing not found in new", field)
		}
		newSchema := n[field]

		// Copy the schemas and remove any child properties for comparison.
		// In theory this will focus in on detecting changes for only the
		// field we are looking at and ignore changes in the children fields.
		// Since we are iterating through the map that should have all fields
		// we should still detect changes in the children fields.
		oldCopy := schema.DeepCopy()
		newCopy := newSchema.DeepCopy()
		oldCopy.Properties = nil
		newCopy.Properties = nil
		if !reflect.DeepEqual(oldCopy, newCopy) {
			diffMap[field] = FieldDiff{
				Old: oldCopy,
				New: newCopy,
			}
		}
	}
	return diffMap, nil
}
