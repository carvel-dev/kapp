// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
)

type EmptyFieldMatcher struct {
	Path Path
}

var _ ResourceMatcher = EmptyFieldMatcher{}

func (m EmptyFieldMatcher) Matches(res Resource) bool {
	return m.check(res.unstructured().Object, m.Path)
}

func (m EmptyFieldMatcher) check(obj interface{}, path Path) bool {
	for i, part := range path {
		switch {
		case part.MapKey != nil:
			typedObj, ok := obj.(map[string]interface{})
			if !ok {
				return obj == nil
			}

			var found bool
			obj, found = typedObj[*part.MapKey]
			if !found {
				// It's not found, so it must be empty
				return true
			}

		case part.ArrayIndex != nil:
			switch {
			case part.ArrayIndex.All != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return obj == nil
				}

				for _, obj := range typedObj {
					empty := m.check(obj, path[i+1:])
					if !empty {
						return false
					}
				}

				return true

			case part.ArrayIndex.Index != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return obj == nil
				}

				if *part.ArrayIndex.Index < len(typedObj) {
					obj = typedObj[*part.ArrayIndex.Index]
				} else {
					// Index not found, it's empty
					return true
				}

			default:
				panic(fmt.Sprintf("Unknown array index: %#v", part.ArrayIndex))
			}

		default:
			panic(fmt.Sprintf("Unexpected path part: %#v", part))
		}
	}

	switch typedObj := obj.(type) {
	case nil:
		return true
	case []interface{}:
		return len(typedObj) == 0
	case map[string]interface{}:
		return len(typedObj) == 0
	default:
		return false
	}
}
