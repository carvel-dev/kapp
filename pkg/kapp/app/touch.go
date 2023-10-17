// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

type Touch struct {
	App              App
	Description      string
	Namespaces       []string
	IgnoreSuccessErr bool
}

func (t Touch) Do(doFunc func() error) error {
	meta := ChangeMeta{
		Description: t.Description,
		Namespaces:  t.Namespaces,
	}

	change, err := t.App.BeginChange(meta)
	if err != nil {
		return err
	}

	workErr := doFunc()
	if workErr != nil {
		_ = change.Fail()
		return workErr
	}

	successErr := change.Succeed()
	if successErr != nil {
		if !t.IgnoreSuccessErr {
			return successErr
		}
	}

	return nil
}
