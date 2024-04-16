// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety_test

import (
	"errors"
	"testing"

	"carvel.dev/kapp/pkg/kapp/crdupgradesafety"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	for _, tc := range []struct {
		name            string
		changeValidator *crdupgradesafety.ChangeValidator
		old             v1.CustomResourceDefinition
		new             v1.CustomResourceDefinition
		shouldError     bool
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
			shouldError: true,
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
			shouldError: true,
		},
		{
			name: "changes, validation failed, change not fully handled, error",
			changeValidator: &crdupgradesafety.ChangeValidator{
				Validations: []crdupgradesafety.ChangeValidation{
					func(_ crdupgradesafety.FieldDiff) (bool, error) {
						return false, errors.New("fail")
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
			shouldError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.changeValidator.Validate(tc.old, tc.new)
			assert.Equal(t, tc.shouldError, err != nil, "should error? - %v", tc.shouldError)
		})
	}
}
