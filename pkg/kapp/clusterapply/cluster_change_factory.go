// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	ctldiff "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diff"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

type ClusterChangeFactory struct {
	opts                ClusterChangeOpts
	identifiedResources ctlres.IdentifiedResources
	changeFactory       ctldiff.ChangeFactory
	changeSetFactory    ctldiff.ChangeSetFactory
	convergedResFactory ConvergedResourceFactory
	ui                  UI
	diffMaskRules       []ctlconf.DiffMaskRule
}

func NewClusterChangeFactory(
	opts ClusterChangeOpts,
	identifiedResources ctlres.IdentifiedResources,
	changeFactory ctldiff.ChangeFactory,
	changeSetFactory ctldiff.ChangeSetFactory,
	convergedResFactory ConvergedResourceFactory,
	ui UI, diffMaskRules []ctlconf.DiffMaskRule,
) ClusterChangeFactory {
	return ClusterChangeFactory{opts, identifiedResources,
		changeFactory, changeSetFactory, convergedResFactory, ui, diffMaskRules}
}

func (f ClusterChangeFactory) NewClusterChange(change ctldiff.Change) *ClusterChange {
	return NewClusterChange(change, f.opts, f.identifiedResources,
		f.changeFactory, f.changeSetFactory, f.convergedResFactory, f.ui, f.diffMaskRules)
}
