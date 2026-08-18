// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety_test

import (
	"errors"
	"fmt"
	"testing"

	"carvel.dev/kapp/pkg/kapp/crdupgradesafety"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
)

func TestEnumChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
					},
				},
				New: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
					},
				},
			},
			shouldHandle: true,
		},
		{
			name: "enum added, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
					},
				},
				New: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
						{
							Raw: []byte("bar"),
						},
					},
				},
			},
			shouldHandle: true,
		},
		{
			name: "no enums before, enums added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
					},
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "enum removed, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
						{
							Raw: []byte("bar"),
						},
					},
				},
				New: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("bar"),
						},
					},
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no enum change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
					},
					ID: "bar",
				},
				New: &v1.JSONSchemaProps{
					Enum: []v1.JSON{
						{
							Raw: []byte("foo"),
						},
					},
					ID: "baz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.EnumChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.Enum)
			assert.Empty(t, tc.diff.New.Enum)
		})
	}
}

func TestCalculateFlatSchemaDiff(t *testing.T) {
	for _, tc := range []struct {
		name         string
		old          crdupgradesafety.FlatSchema
		new          crdupgradesafety.FlatSchema
		expectedDiff map[string]crdupgradesafety.FieldDiff
		shouldError  bool
	}{
		{
			name: "no diff in schemas, empty diff, no error",
			old: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{},
			},
			new: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{},
			},
			expectedDiff: map[string]crdupgradesafety.FieldDiff{},
		},
		{
			name: "diff in schemas, diff returned, no error",
			old: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{},
			},
			new: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{
					ID: "bar",
				},
			},
			expectedDiff: map[string]crdupgradesafety.FieldDiff{
				"foo": {
					Old: &v1.JSONSchemaProps{},
					New: &v1.JSONSchemaProps{ID: "bar"},
				},
			},
		},
		{
			name: "diff in child properties only, no diff returned, no error",
			old: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"bar": {ID: "bar"},
					},
				},
			},
			new: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{
					Properties: map[string]v1.JSONSchemaProps{
						"bar": {ID: "baz"},
					},
				},
			},
			expectedDiff: map[string]crdupgradesafety.FieldDiff{},
		},
		{
			name: "field exists in old but not new, no diff returned, error",
			old: crdupgradesafety.FlatSchema{
				"foo": &v1.JSONSchemaProps{},
			},
			new: crdupgradesafety.FlatSchema{
				"bar": &v1.JSONSchemaProps{},
			},
			expectedDiff: map[string]crdupgradesafety.FieldDiff{},
			shouldError:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			diff, err := crdupgradesafety.CalculateFlatSchemaDiff(tc.old, tc.new)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.expectedDiff, diff)
		})
	}
}

func TestFlattenSchema(t *testing.T) {
	schema := &v1.JSONSchemaProps{
		Properties: map[string]v1.JSONSchemaProps{
			"foo": {
				Properties: map[string]v1.JSONSchemaProps{
					"bar": {},
				},
			},
			"baz": {},
		},
	}

	foo := schema.Properties["foo"]
	foobar := schema.Properties["foo"].Properties["bar"]
	baz := schema.Properties["baz"]
	expected := crdupgradesafety.FlatSchema{
		"^":         schema,
		"^.foo":     &foo,
		"^.foo.bar": &foobar,
		"^.baz":     &baz,
	}

	actual := crdupgradesafety.FlattenSchema(schema)

	assert.Equal(t, expected, actual)
}

func TestChangeValidator(t *testing.T) {
	validationErr1 := errors.New(`version "v1alpha1", field "^" has unknown change, refusing to determine that change is safe`)
	validationErr2 := errors.New(`version "v1alpha1", field "^": fail`)

	for _, tc := range []struct {
		name            string
		changeValidator *crdupgradesafety.ChangeValidator
		old             v1.CustomResourceDefinition
		new             v1.CustomResourceDefinition
		expectedError   error
	}{
		{
			name: "no changes, no error",
			changeValidator: &crdupgradesafety.ChangeValidator{
				Validations: []crdupgradesafety.ChangeValidation{
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return false, errors.New("should not run")
					},
				},
			},
			old: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			new: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{},
							},
						},
					},
				},
			},
		},
		{
			name: "changes, validation successful, change is fully handled, no error",
			changeValidator: &crdupgradesafety.ChangeValidator{
				Validations: []crdupgradesafety.ChangeValidation{
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return true, nil
					},
				},
			},
			old: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			new: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{
									ID: "foo",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "changes, validation successful, change not fully handled, error",
			changeValidator: &crdupgradesafety.ChangeValidator{
				Validations: []crdupgradesafety.ChangeValidation{
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return false, nil
					},
				},
			},
			old: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			new: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			expectedError: validationErr1,
		},
		{
			name: "changes, validation failed, change fully handled, error",
			changeValidator: &crdupgradesafety.ChangeValidator{
				Validations: []crdupgradesafety.ChangeValidation{
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return true, errors.New("fail")
					},
				},
			},
			old: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			new: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			expectedError: validationErr2,
		},
		{
			name: "changes, validation failed, change not fully handled, ordered error",
			changeValidator: &crdupgradesafety.ChangeValidator{
				Validations: []crdupgradesafety.ChangeValidation{
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return false, errors.New("fail")
					},
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return false, errors.New("error")
					},
				},
			},
			old: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{},
							},
						},
					},
				},
			},
			new: v1.CustomResourceDefinition{
				Spec: v1.CustomResourceDefinitionSpec{
					Versions: []v1.CustomResourceDefinitionVersion{
						{
							Name: "v1alpha1",
							Schema: &v1.CustomResourceValidation{
								OpenAPIV3Schema: &v1.JSONSchemaProps{
									ID: "foo",
								},
							},
						},
					},
				},
			},
			expectedError: fmt.Errorf("%w\n%s\n%w", validationErr2, `version "v1alpha1", field "^": error`, validationErr1),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.changeValidator.Validate(tc.old, tc.new)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequiredFieldChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Required: []string{"foo"},
				},
				New: &v1.JSONSchemaProps{
					Required: []string{"foo"},
				},
			},
			shouldHandle: true,
		},
		{
			name: "required field removed, no other changes, should be handled, no error",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Required: []string{"foo", "bar"},
				},
				New: &v1.JSONSchemaProps{
					Required: []string{"foo"},
				},
			},
			shouldHandle: true,
		},
		{
			name: "new required field added, no other changes, should be handled, error",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Required: []string{"foo"},
				},
				New: &v1.JSONSchemaProps{
					Required: []string{"foo", "bar"},
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no required field change, another field modified, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Required: []string{"foo"},
					ID:       "abc",
				},
				New: &v1.JSONSchemaProps{
					Required: []string{"foo"},
					ID:       "xyz",
				},
			},
		},
		{
			name: "no required fields before, new required fields added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					Required: []string{"foo", "bar"},
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.RequiredFieldChangeValidation(tc.diff)
			assert.Empty(t, tc.diff.Old.Required)
			assert.Empty(t, tc.diff.New.Required)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
		})
	}
}

func TestMinimumChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(10),
				},
				New: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum decreased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(10),
				},
				New: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(8),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum increased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(8),
				},
				New: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no minimum before, minimum added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(8),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "minimum removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(8),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no minimum change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(8),
					ID:      "bar",
				},
				New: &v1.JSONSchemaProps{
					Minimum: pointer.Float64(8),
					ID:      "baz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MinimumChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.Minimum)
			assert.Empty(t, tc.diff.New.Minimum)
		})
	}
}

func TestMinimumLengthChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum length decreased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(8),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum length increased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(8),
				},
				New: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no minimum length before, minimum length added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "minimum length removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no minimum length change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
					ID:        "bar",
				},
				New: &v1.JSONSchemaProps{
					MinLength: pointer.Int64(10),
					ID:        "baz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MinimumLengthChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.MinLength)
			assert.Empty(t, tc.diff.New.MinLength)
		})
	}
}

func TestMinimumItemsChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum items decreased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(8),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum items increased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(8),
				},
				New: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no minimum items before, minimum items added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "minimum items removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no minimum items change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
					ID:       "bar",
				},
				New: &v1.JSONSchemaProps{
					MinItems: pointer.Int64(10),
					ID:       "baz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MinimumItemsChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.MinItems)
			assert.Empty(t, tc.diff.New.MinItems)
		})
	}
}

func TestMinimumPropertiesChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum properties decreased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(8),
				},
			},
			shouldHandle: true,
		},
		{
			name: "minimum properties increased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(8),
				},
				New: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no minimum properties before, minimum properties added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "minimum properties removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no minimum properties change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
					ID:            "bar",
				},
				New: &v1.JSONSchemaProps{
					MinProperties: pointer.Int64(10),
					ID:            "baz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MinimumPropertiesChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.MinProperties)
			assert.Empty(t, tc.diff.New.MinProperties)
		})
	}
}

func TestMaximumChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
				},
				New: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum increased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
				},
				New: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(100),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum decreased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
				},
				New: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(1),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no maximum before, maximum added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "maximum removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no maximum change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
					ID:      "abc",
				},
				New: &v1.JSONSchemaProps{
					Maximum: pointer.Float64(10),
					ID:      "xyz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MaximumChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.Maximum)
			assert.Empty(t, tc.diff.New.Maximum)
		})
	}
}

func TestMaximumLengthChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum length increased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(100),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum length decreased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(1),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no maximum length before, maximum length added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "maximum length removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no maximum length change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
					ID:        "abc",
				},
				New: &v1.JSONSchemaProps{
					MaxLength: pointer.Int64(10),
					ID:        "xyz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MaximumLengthChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.MaxLength)
			assert.Empty(t, tc.diff.New.MaxLength)
		})
	}
}

func TestMaximumItemsChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum items increased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(100),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum items decreased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(1),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no maximum items before, maximum items added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "maximum items removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no maximum items change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
					ID:       "abc",
				},
				New: &v1.JSONSchemaProps{
					MaxItems: pointer.Int64(10),
					ID:       "xyz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MaximumItemsChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.MaxItems)
			assert.Empty(t, tc.diff.New.MaxItems)
		})
	}
}

func TestMaximumPropertiesChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum properties increased, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(100),
				},
			},
			shouldHandle: true,
		},
		{
			name: "maximum properties decreased, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(1),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no maximum properties before, maximum properties added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "maximum properties removed, no other changes, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
		},
		{
			name: "no maximum properties change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
					ID:            "abc",
				},
				New: &v1.JSONSchemaProps{
					MaxProperties: pointer.Int64(10),
					ID:            "xyz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.MaximumPropertiesChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.MaxProperties)
			assert.Empty(t, tc.diff.New.MaxProperties)
		})
	}
}

func TestDefaultChangeValidation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		diff         crdupgradesafety.FieldDiff
		shouldError  bool
		shouldHandle bool
	}{
		{
			name: "no change in default value, no error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
				},
				New: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
				},
			},
			shouldHandle: true,
		},
		{
			name: "no default before, default added, no other changes, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{},
				New: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "existing default removed, no other changes, error, should be handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
				},
				New: &v1.JSONSchemaProps{},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "default value changed, error, marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
				},
				New: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("bar"),
					},
				},
			},
			shouldHandle: true,
			shouldError:  true,
		},
		{
			name: "no default value change, other changes, no error, not marked as handled",
			diff: crdupgradesafety.FieldDiff{
				Old: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
					ID: "abc",
				},
				New: &v1.JSONSchemaProps{
					Default: &v1.JSON{
						Raw: []byte("foo"),
					},
					ID: "xyz",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := crdupgradesafety.DefaultValueChangeValidation(tc.diff)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
			assert.Equal(t, tc.shouldHandle, handled, "should be handled? - %v", tc.shouldHandle)
			assert.Empty(t, tc.diff.Old.Default)
			assert.Empty(t, tc.diff.New.Default)
		})
	}
}
