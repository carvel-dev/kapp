// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"
	"sort"

	"github.com/k14s/ytt/pkg/filepos"

	"github.com/k14s/ytt/pkg/structmeta"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

const (
	AnnotationNullable     structmeta.AnnotationName = "schema/nullable"
	AnnotationType         structmeta.AnnotationName = "schema/type"
	TypeAnnotationKwargAny string                    = "any"
)

type Annotation interface {
	NewTypeFromAnn() yamlmeta.Type
}

type TypeAnnotation struct {
	any          bool
	itemPosition *filepos.Position
}

type NullableAnnotation struct {
	providedValueType yamlmeta.Type
	itemPosition      *filepos.Position
}

func NewTypeAnnotation(ann template.NodeAnnotation, pos *filepos.Position) (*TypeAnnotation, error) {
	if len(ann.Kwargs) == 0 {
		return nil, fmt.Errorf("expected @%v annotation to have keyword argument and value. Supported key-value pairs are '%v=True', '%v=False'", AnnotationType, TypeAnnotationKwargAny, TypeAnnotationKwargAny)
	}
	typeAnn := &TypeAnnotation{itemPosition: pos}
	for _, kwarg := range ann.Kwargs {
		argName, err := core.NewStarlarkValue(kwarg[0]).AsString()
		if err != nil {
			return nil, err
		}

		switch argName {
		case TypeAnnotationKwargAny:
			isAnyType, err := core.NewStarlarkValue(kwarg[1]).AsBool()
			if err != nil {
				return nil,
					fmt.Errorf("processing @%v '%v' argument: %s", AnnotationType, TypeAnnotationKwargAny, err)
			}
			typeAnn.any = isAnyType

		default:
			return nil, fmt.Errorf("unknown @%v annotation keyword argument '%v'. Supported kwargs are '%v'", AnnotationType, argName, TypeAnnotationKwargAny)
		}
	}
	return typeAnn, nil
}

func NewNullableAnnotation(ann template.NodeAnnotation, valueType yamlmeta.Type, pos *filepos.Position) (*NullableAnnotation, error) {
	if len(ann.Kwargs) != 0 {
		return nil, fmt.Errorf("expected @%v annotation to not contain any keyword arguments", AnnotationNullable)
	}

	return &NullableAnnotation{valueType, pos}, nil
}

func (t *TypeAnnotation) NewTypeFromAnn() yamlmeta.Type {
	if t.any {
		return &AnyType{Position: t.itemPosition}
	}
	return nil
}

func (t *TypeAnnotation) IsAny() bool {
	return t.any
}

func (n *NullableAnnotation) NewTypeFromAnn() yamlmeta.Type {
	return &NullType{ValueType: n.providedValueType, Position: n.itemPosition}
}

func collectAnnotations(item yamlmeta.Node) ([]Annotation, error) {
	var anns []Annotation

	for _, annotation := range []structmeta.AnnotationName{AnnotationType, AnnotationNullable} {
		ann, err := processOptionalAnnotation(item, annotation)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}
	return anns, nil
}

func processOptionalAnnotation(node yamlmeta.Node, optionalAnnotation structmeta.AnnotationName) (Annotation, error) {
	nodeAnnotations := template.NewAnnotations(node)

	if nodeAnnotations.Has(optionalAnnotation) {
		ann, _ := nodeAnnotations[optionalAnnotation]

		switch optionalAnnotation {
		case AnnotationNullable:
			wrappedValueType, err := newCollectionItemValueType(node.GetValues()[0], node.GetPosition())
			if err != nil {
				return nil, err
			}
			nullAnn, err := NewNullableAnnotation(ann, wrappedValueType, node.GetPosition())
			if err != nil {
				return nil, NewInvalidSchemaError(node, err.Error(), "")
			}
			return nullAnn, nil
		case AnnotationType:
			typeAnn, err := NewTypeAnnotation(ann, node.GetPosition())
			if err != nil {
				return nil, NewInvalidSchemaError(node, err.Error(), "")
			}
			return typeAnn, nil
		}
	}

	return nil, nil
}

func convertAnnotationsToSingleType(anns []Annotation) yamlmeta.Type {
	annsCopy := append([]Annotation{}, anns...)

	if len(annsCopy) == 0 {
		return nil
	}

	// allow Configuration Author to annotate "nullable" as a fallback if "any" is false.
	preferAnyTypeOverNullableType := func(i, j int) bool {
		if typeAnn, ok := annsCopy[i].(*TypeAnnotation); ok && typeAnn.IsAny() {
			return true
		}
		return false
	}

	sort.Slice(annsCopy, preferAnyTypeOverNullableType)
	return annsCopy[0].NewTypeFromAnn()
}
