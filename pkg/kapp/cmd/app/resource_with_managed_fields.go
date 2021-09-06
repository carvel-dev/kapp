// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ResourceWithManagedFields struct {
	res           ctlres.Resource
	managedFields bool
}

func NewResourceWithManagedFields(res ctlres.Resource, managedFields bool) ResourceWithManagedFields {
	if res != nil {
		res = res.DeepCopy()
	}
	return ResourceWithManagedFields{res: res, managedFields: managedFields}
}

func (r ResourceWithManagedFields) Resource() (ctlres.Resource, error) {
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

func (ResourceWithManagedFields) removeManagedFieldsResMods() ctlres.ResourceMod {
	return ctlres.FieldRemoveMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "managedFields"}),
	}
}
