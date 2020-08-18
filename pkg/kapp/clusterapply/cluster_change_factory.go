// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ClusterChangeFactory struct {
	opts                ClusterChangeOpts
	identifiedResources ctlres.IdentifiedResources
	changeFactory       ctldiff.ChangeFactory
	changeSetFactory    ctldiff.ChangeSetFactory
	convergedResFactory ConvergedResourceFactory
	ui                  UI
}

func NewClusterChangeFactory(
	opts ClusterChangeOpts,
	identifiedResources ctlres.IdentifiedResources,
	changeFactory ctldiff.ChangeFactory,
	changeSetFactory ctldiff.ChangeSetFactory,
	convergedResFactory ConvergedResourceFactory, ui UI,
) ClusterChangeFactory {
	return ClusterChangeFactory{opts, identifiedResources,
		changeFactory, changeSetFactory, convergedResFactory, ui}
}

func (f ClusterChangeFactory) NewClusterChange(change ctldiff.Change) *ClusterChange {
	return NewClusterChange(change, f.opts, f.identifiedResources,
		f.changeFactory, f.changeSetFactory, f.convergedResFactory, f.ui)
}
