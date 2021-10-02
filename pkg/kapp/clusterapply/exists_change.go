// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ExistsChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
}

func (c ExistsChange) ApplyStrategy() (ApplyStrategy, error) {
	res := c.change.NewResource()
	return ExistsStrategy{res, c.identifiedResources}, nil
}

type ExistsStrategy struct {
	res                 ctlres.Resource
	identifiedResources ctlres.IdentifiedResources
}

func (e ExistsStrategy) Op() ClusterChangeApplyStrategyOp { return "" }

func (e ExistsStrategy) Apply() error {
	exists, err := e.identifiedResources.Exists(e.res, ctlres.ExistsOpts{})
	if !exists {
		if err != nil {
			return err
		}
		return ExistsChangeError{}
	}
	return nil
}

type ExistsChangeError struct{}

func (e ExistsChangeError) Error() string {
	return fmt.Sprint("External resource doesn't exists")
}
