// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ManagedFieldsResource struct {
	res ctlres.Resource
	managedFields bool
}

func NewManagedFieldsResource(res ctlres.Resource, managedFields bool) ManagedFieldsResource {
	if res != nil {
		res = res.DeepCopy()
	}
	return ManagedFieldsResource{res: res, managedFields: managedFields}
}

func (r ManagedFieldsResource) Resource() (ctlres.Resource, error) {
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

func (ManagedFieldsResource) removeManagedFieldsResMods() ctlres.ResourceMod {
	return ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "managedFields"}),
	}
}
