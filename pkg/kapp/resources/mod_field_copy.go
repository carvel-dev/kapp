// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
)

type FieldCopyModSource string

const (
	FieldCopyModSourceNew      FieldCopyModSource = "new"
	FieldCopyModSourceExisting                    = "existing"
)

type FieldCopyMod struct {
	ResourceMatcher ResourceMatcher
	Path            Path
	Sources         []FieldCopyModSource // first preferred
}

var _ ResourceModWithMultiple = FieldCopyMod{}

func (t FieldCopyMod) IsResourceMatching(res Resource) bool {
	if res == nil || !t.ResourceMatcher.Matches(res) {
		return false
	}
	return true
}

func (t FieldCopyMod) ApplyFromMultiple(res Resource, srcs map[FieldCopyModSource]Resource) error {
	// Make a copy of resource, to avoid modifications
	// that may be done even in case when there is nothing to copy
	updatedRes := res.DeepCopy()

	updated, err := t.apply(updatedRes.unstructured().Object, t.Path, Path{}, srcs)
	if err != nil {
		return fmt.Errorf("FieldCopyMod for path '%s' on resource '%s': %s", t.Path.AsString(), res.Description(), err)
	}

	if updated {
		res.setUnstructured(updatedRes.unstructured())
	}

	return nil
}

func (t FieldCopyMod) apply(obj interface{}, path Path, fullPath Path, srcs map[FieldCopyModSource]Resource) (bool, error) {
	for i, part := range path {
		isLast := len(path) == i+1
		fullPath = append(fullPath, part)

		switch {
		case part.MapKey != nil:
			typedObj, ok := obj.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("Unexpected non-map found: %T", obj)
			}

			if isLast {
				return t.copyIntoMap(typedObj, fullPath, srcs)
			}

			var found bool
			obj, found = typedObj[*part.MapKey]
			// TODO check strictness?
			if !found || obj == nil {
				// create empty maps if there are no downstream array indexes;
				// if there are, we cannot make them anyway, so just exit
				if path.ContainsNonMapKeys() {
					return false, nil
				}
				obj = map[string]interface{}{}
				typedObj[*part.MapKey] = obj
			}

		case part.ArrayIndex != nil:
			if isLast {
				return false, fmt.Errorf("Expected last part of the path to be map key")
			}

			switch {
			case part.ArrayIndex.All != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return false, fmt.Errorf("Unexpected non-array found: %T", obj)
				}

				var anyUpdated bool

				for objI, obj := range typedObj {
					objI := objI

					newFullPath := append([]*PathPart{}, fullPath...)
					newFullPath[len(newFullPath)-1] = &PathPart{ArrayIndex: &PathPartArrayIndex{Index: &objI}}

					updated, err := t.apply(obj, path[i+1:], newFullPath, srcs)
					if err != nil {
						return false, err
					}
					if updated {
						anyUpdated = true
					}
				}

				return anyUpdated, nil // dealt with children, get out

			case part.ArrayIndex.Index != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return false, fmt.Errorf("Unexpected non-array found: %T", obj)
				}

				if *part.ArrayIndex.Index < len(typedObj) {
					obj = typedObj[*part.ArrayIndex.Index]
					return t.apply(obj, path[i+1:], fullPath, srcs)
				}

				return false, nil // index not found, nothing to append to

			default:
				panic(fmt.Sprintf("Unknown array index: %#v", part.ArrayIndex))
			}

		default:
			panic(fmt.Sprintf("Unexpected path part: %#v", part))
		}
	}

	panic("unreachable")
}

func (t FieldCopyMod) copyIntoMap(obj map[string]interface{}, fullPath Path, srcs map[FieldCopyModSource]Resource) (bool, error) {
	lastPartPath := fullPath[len(fullPath)-1]
	if lastPartPath.MapKey == nil {
		return false, fmt.Errorf("Expected last path part to be map-key")
	}

	for _, src := range t.Sources {
		srcRes, found := srcs[src]
		if !found || srcRes == nil {
			continue
		}

		val, found, err := t.obtainValue(srcRes.unstructured().Object, fullPath)
		if err != nil {
			return false, err
		} else if !found {
			continue
		}

		obj[*lastPartPath.MapKey] = val
		return true, nil
	}

	return false, nil
}

func (t FieldCopyMod) obtainValue(obj interface{}, path Path) (interface{}, bool, error) {
	for i, part := range path {
		isLast := len(path) == i+1

		switch {
		case part.MapKey != nil:
			typedObj, ok := obj.(map[string]interface{})
			if !ok && typedObj != nil {
				return nil, false, fmt.Errorf("Unexpected non-map found: %T", obj)
			}

			var found bool
			obj, found = typedObj[*part.MapKey]
			if !found {
				return nil, false, nil // index is not found return
			}

		case part.ArrayIndex != nil:
			if isLast {
				return nil, false, fmt.Errorf("Expected last part of the path to be map key")
			}

			switch {
			case part.ArrayIndex.Index != nil:
				typedObj, ok := obj.([]interface{})
				if !ok {
					return nil, false, fmt.Errorf("Unexpected non-array found: %T", obj)
				}

				if *part.ArrayIndex.Index < len(typedObj) {
					obj = typedObj[*part.ArrayIndex.Index]
				} else {
					return nil, false, nil // index not found, return
				}

			default:
				panic(fmt.Sprintf("Unknown array index: %#v", part.ArrayIndex))
			}

		default:
			panic(fmt.Sprintf("Unexpected path part: %#v", part))
		}
	}

	return obj, true, nil
}
