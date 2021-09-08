// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"encoding/json"
	"fmt"
	"strings"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/apimachinery/pkg/types"
)

const (
	deleteStrategyAnnKey                                      = "kapp.k14s.io/delete-strategy"
	deleteStrategyPlainAnnValue  ClusterChangeApplyStrategyOp = ""
	deleteStrategyOrphanAnnValue ClusterChangeApplyStrategyOp = "orphan"

	appLabelKey      = "kapp.k14s.io/app" // TODO duplicated here
	orphanedLabelKey = "kapp.k14s.io/orphaned"
)

var (
	jsonPointerEncoder = strings.NewReplacer("~", "~0", "/", "~1")
)

type DeleteChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
}

func (c DeleteChange) ApplyStrategy() (ApplyStrategy, error) {
	res := c.change.ExistingResource()
	strategy := res.Annotations()[deleteStrategyAnnKey]

	switch ClusterChangeApplyStrategyOp(strategy) {
	case deleteStrategyPlainAnnValue:
		return DeletePlainStrategy{res, c}, nil

	case deleteStrategyOrphanAnnValue:
		return DeleteOrphanStrategy{res, c}, nil

	default:
		return nil, fmt.Errorf("Unknown delete strategy: %s", strategy)
	}
}

func (c DeleteChange) IsDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	res := c.change.ExistingResource()

	switch ClusterChangeApplyStrategyOp(res.Annotations()[deleteStrategyAnnKey]) {
	case deleteStrategyOrphanAnnValue:
		return ctlresm.DoneApplyState{Done: true, Successful: true, Message: "Resource orphaned"}, nil, nil
	}

	// it should not matter if change is ignored or not
	// because it should be deleted eventually anyway (thru GC)
	// We should check for the UID check because of the following bug:
	// https://github.com/vmware-tanzu/carvel-kapp/issues/229
	exists, err := c.identifiedResources.Exists(res, ctlres.ExistsOpts{SameUID: true})
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	return ctlresm.DoneApplyState{Done: !exists, Successful: true}, nil, nil
}

type DeletePlainStrategy struct {
	res ctlres.Resource
	d   DeleteChange
}

func (c DeletePlainStrategy) Op() ClusterChangeApplyStrategyOp { return deleteStrategyPlainAnnValue }

func (c DeletePlainStrategy) Apply() error {
	// TODO should we be configuring default garbage collection policy to background?
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
	return c.d.identifiedResources.Delete(c.res)
}

type DeleteOrphanStrategy struct {
	res ctlres.Resource
	d   DeleteChange
}

func (c DeleteOrphanStrategy) Op() ClusterChangeApplyStrategyOp { return deleteStrategyOrphanAnnValue }

func (c DeleteOrphanStrategy) Apply() error {
	mergePatch := []interface{}{
		// TODO currently we do not account for when '-a label:foo=bar' used
		map[string]interface{}{
			"op":   "remove",
			"path": "/metadata/labels/" + jsonPointerEncoder.Replace(appLabelKey),
		},
		map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + jsonPointerEncoder.Replace(orphanedLabelKey),
			"value": "",
		},
	}

	patchJSON, err := json.Marshal(mergePatch)
	if err != nil {
		return err
	}

	_, err = c.d.identifiedResources.Patch(c.res, types.JSONPatchType, patchJSON)
	return err
}
