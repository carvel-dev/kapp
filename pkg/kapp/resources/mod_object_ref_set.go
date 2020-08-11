// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
)

type ObjectRefSetMod struct {
	ResourceMatcher ResourceMatcher
	Path            Path
	ReplacementFunc func(map[string]interface{}) error
}

var _ ResourceMod = ObjectRefSetMod{}

func (t ObjectRefSetMod) Apply(res Resource) error {
	if !t.ResourceMatcher.Matches(res) {
		return nil
	}
	err := t.apply(res.unstructured().Object, t.Path)
	if err != nil {
		return fmt.Errorf("ObjectRefSetMod for path '%s' on resource '%s': %s", t.Path.AsString(), res.Description(), err)
	}
	return nil
}

func (t ObjectRefSetMod) apply(obj interface{}, path Path) error {
	for i, part := range path {
		switch {
		case part.MapKey != nil:
			typedObj, ok := obj.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Unexpected non-map found: %T", obj)
			}

			var found bool
			obj, found = typedObj[*part.MapKey]
			if !found {
				return nil
			}

		case part.ArrayIndex != nil:
			switch {
			case part.ArrayIndex.All != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return fmt.Errorf("Unexpected non-array found: %T", obj)
				}

				for _, obj := range typedObj {
					err := t.apply(obj, path[i+1:])
					if err != nil {
						return err
					}
				}

				return nil // dealt with children, get out

			case part.ArrayIndex.Index != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return fmt.Errorf("Unexpected non-array found: %T", obj)
				}

				if *part.ArrayIndex.Index < len(typedObj) {
					return t.apply(typedObj[*part.ArrayIndex.Index], path[i+1:])
				}

				return nil // index not found, nothing to append to

			default:
				panic(fmt.Sprintf("Unknown array index: %#v", part.ArrayIndex))
			}

		default:
			panic(fmt.Sprintf("Unexpected path part: %#v", part))
		}
	}

	typedObj, ok := obj.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Unexpected non-map found: %T", obj)
	}

	return t.ReplacementFunc(typedObj)
}
