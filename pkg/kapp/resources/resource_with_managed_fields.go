// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

type ResourceWithManagedFields struct {
	res           Resource
	managedFields bool
}

func NewResourceWithManagedFields(res Resource, managedFields bool) ResourceWithManagedFields {
	return ResourceWithManagedFields{res: res, managedFields: managedFields}
}

func (r ResourceWithManagedFields) Resource() (Resource, error) {
	res := r.res.DeepCopy()
	if r.managedFields {
		return res, nil
	}
	err := r.removeManagedFieldsResMods().Apply(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ResourceWithManagedFields) removeManagedFieldsResMods() ResourceMod {
	return FieldRemoveMod{
		ResourceMatcher: AllMatcher{},
		Path:            NewPathFromStrings([]string{"metadata", "managedFields"}),
	}
}
