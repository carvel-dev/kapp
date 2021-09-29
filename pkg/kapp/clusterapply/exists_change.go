// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"errors"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ExistsChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
}

func (c ExistsChange) ApplyStrategy() (ApplyStrategy, error) {
	res := c.change.NewResource()
	return WaitStrategy{res, c}, nil
}

type WaitStrategy struct {
	res ctlres.Resource
	e   ExistsChange
}

func (c WaitStrategy) Op() ClusterChangeApplyStrategyOp { return "" }

func (c WaitStrategy) Apply() error {
	exists, err := c.e.identifiedResources.Exists(c.res, ctlres.ExistsOpts{})
	if !exists {
		if err != nil {
			return err
		}
		return errors.New("Placeholder resource doesn't exists")
	}
	return nil
}
