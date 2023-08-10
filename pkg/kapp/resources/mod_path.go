// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ResourceMod interface {
	Apply(Resource) error
}

type ResourceModWithMultiple interface {
	ApplyFromMultiple(Resource, map[FieldCopyModSource]Resource) error
}

type Path []*PathPart

type PathPart struct {
	MapKey     *string
	ArrayIndex *PathPartArrayIndex
}

var _ json.Unmarshaler = &PathPart{}

type PathPartArrayIndex struct {
	Index *int
	All   *bool `json:"allIndexes"`
}

func NewPathFromStrings(strs []string) Path {
	var path Path
	for _, str := range strs {
		path = append(path, NewPathPartFromString(str))
	}
	return path
}

func NewPathFromInterfaces(parts []interface{}) Path {
	var path Path
	for _, part := range parts {
		switch typedPart := part.(type) {
		case string:
			path = append(path, NewPathPartFromString(typedPart))
		case int:
			path = append(path, NewPathPartFromIndex(typedPart))
		default:
			panic(fmt.Sprintf("Unexpected part type %T", typedPart))
		}
	}
	return path
}

func (p Path) AsStrings() []string {
	var result []string
	for _, part := range p {
		if part.MapKey == nil {
			panic(fmt.Sprintf("Unexpected non-map-key path part '%#v'", part))
		}
		result = append(result, *part.MapKey)
	}
	return result
}

func (p Path) AsString() string {
	var result []string
	for _, part := range p {
		result = append(result, part.AsString())
	}
	return strings.Join(result, ",")
}

func (p Path) ContainsNonMapKeys() bool {
	for _, part := range p {
		if part.MapKey == nil {
			return true
		}
	}
	return false
}

func NewPathPartFromString(str string) *PathPart {
	return &PathPart{MapKey: &str}
}

func NewPathPartFromIndex(i int) *PathPart {
	return &PathPart{ArrayIndex: &PathPartArrayIndex{Index: &i}}
}

func NewPathPartFromIndexAll() *PathPart {
	trueBool := true
	return &PathPart{ArrayIndex: &PathPartArrayIndex{All: &trueBool}}
}

func (p *PathPart) AsString() string {
	switch {
	case p.MapKey != nil:
		return *p.MapKey
	case p.ArrayIndex != nil && p.ArrayIndex.Index != nil:
		return fmt.Sprintf("%d", *p.ArrayIndex.Index)
	case p.ArrayIndex != nil && p.ArrayIndex.All != nil:
		return "(all)"
	default:
		panic("Unknown path part")
	}
}

func (p *PathPart) UnmarshalJSON(data []byte) error {
	var str string
	var idx PathPartArrayIndex

	switch {
	case json.Unmarshal(data, &str) == nil:
		p.MapKey = &str
	case json.Unmarshal(data, &idx) == nil:
		p.ArrayIndex = &idx
	default:
		return fmt.Errorf("Unknown path part")
	}
	return nil
}
