// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
)

type StringMapAppendMod struct {
	// For example applies to Deployment, ReplicaSet, StatefulSet
	// TODO should there be an opt-out way?
	ResourceMatcher ResourceMatcher
	Path            Path
	SkipIfNotFound  bool
	KVs             map[string]string
}

var _ ResourceMod = StringMapAppendMod{}

func (t StringMapAppendMod) Apply(res Resource) error {
	if !t.ResourceMatcher.Matches(res) {
		return nil
	}
	err := t.apply(res.unstructured().Object, t.Path)
	if err != nil {
		return fmt.Errorf("StringMapAppendMod for path '%s' on resource '%s': %s", t.Path.AsString(), res.Description(), err)
	}
	return nil
}

func (t StringMapAppendMod) apply(obj interface{}, path Path) error {
	for i, part := range path {
		switch {
		case part.MapKey != nil:
			typedObj, ok := obj.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Unexpected non-map found: %T", obj)
			}

			var found bool
			obj, found = typedObj[*part.MapKey]
			// TODO check strictness?
			if !found || obj == nil {
				// create empty maps if there are no downstream array indexes;
				// if there are, we cannot make them anyway, so just exit
				if t.SkipIfNotFound || path.ContainsNonMapKeys() {
					return nil
				}
				obj = map[string]interface{}{}
				typedObj[*part.MapKey] = obj
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

	for k, v := range t.KVs {
		typedObj[k] = v
	}

	return nil
}
