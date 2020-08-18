// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ResourceWithRemovedFields struct {
	res  ctlres.Resource
	mods []ctlres.FieldRemoveMod
}

func NewResourceWithRemovedFields(res ctlres.Resource, mods []ctlres.FieldRemoveMod) ResourceWithRemovedFields {
	if res != nil {
		res = res.DeepCopy()
	}
	return ResourceWithRemovedFields{res: res, mods: mods}
}

func (r ResourceWithRemovedFields) Resource() (ctlres.Resource, error) {
	if r.res == nil {
		return nil, nil // nothing to ignore
	}

	result := r.res.DeepCopy()

	for _, t := range r.mods {
		err := t.Apply(result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
