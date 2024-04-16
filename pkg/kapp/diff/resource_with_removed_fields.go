// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type ResourceWithRemovedFields struct {
	res  ctlres.Resource
	mods []ctlres.FieldRemoveMod
}

func NewResourceWithRemovedFields(res ctlres.Resource, mods []ctlres.FieldRemoveMod) ResourceWithRemovedFields {
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
