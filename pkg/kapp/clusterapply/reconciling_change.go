// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	ctldiff "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diff"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/logger"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
	ctlresm "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resourcesmisc"
)

type ReconcilingChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
	convergedResFactory ConvergedResourceFactory
}

type SpecificResource interface {
	IsDoneApplying() ctlresm.DoneApplyState
}

func (c ReconcilingChange) IsDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	labeledResources := ctlres.NewLabeledResources(nil, c.identifiedResources, logger.NewTODOLogger())

	// Refresh resource with latest changes from the server
	// Pick up new or existing resource (and not just new resource),
	// as some changes may be apply->noop, wait->reconcile.
	parentRes, err := c.identifiedResources.Get(c.change.NewOrExistingResource())
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	return c.convergedResFactory.New(parentRes, labeledResources.GetAssociated).IsDoneApplying()
}
