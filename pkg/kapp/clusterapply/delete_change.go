// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"encoding/json"
	"fmt"
	"strings"

	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

type inoperableResourceRef struct {
	schema.GroupKind
	Name string
}

// List of resources use the "orphan" delete strategy implicitly as they cannot be deleted by end users
var inoperableResourceList = []inoperableResourceRef{
	{
		GroupKind: schema.GroupKind{Group: "", Kind: "Namespace"},
		Name:      "default",
	},
	{
		GroupKind: schema.GroupKind{Group: "", Kind: "Namespace"},
		Name:      "kube-node-lease",
	},
	{
		GroupKind: schema.GroupKind{Group: "", Kind: "Namespace"},
		Name:      "kube-public",
	},
	{
		GroupKind: schema.GroupKind{Group: "", Kind: "Namespace"},
		Name:      "kube-system",
	},
}

func (c DeleteChange) ApplyStrategy() (ApplyStrategy, error) {
	res := c.change.ExistingResource()
	strategy := res.Annotations()[deleteStrategyAnnKey]

	if c.isInoperableResource() {
		return DeleteOrphanStrategy{res, c}, nil
	}

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

	if c.isInoperableResource() {
		return ctlresm.DoneApplyState{Done: true, Successful: true, Message: "Resource orphaned"}, nil, nil
	}

	switch ClusterChangeApplyStrategyOp(res.Annotations()[deleteStrategyAnnKey]) {
	case deleteStrategyOrphanAnnValue:
		return ctlresm.DoneApplyState{Done: true, Successful: true, Message: "Resource orphaned"}, nil, nil
	}

	// it should not matter if change is ignored or not
	// because it should be deleted eventually anyway (thru GC)
	// We should check for the UID check because of the following bug:
	// https://github.com/carvel-dev/kapp/issues/229
	existingRes, exists, err := c.identifiedResources.Exists(res, ctlres.ExistsOpts{SameUID: true})
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	if !exists {
		return ctlresm.DoneApplyState{Done: true, Successful: true}, nil, nil
	}

	return ctlresm.DoneApplyState{Done: false, Successful: true}, descMessage(existingRes), nil
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
		// Reason: In labeled app labels are not accessible in delete operation now. Without the label info kapp can not apply the changes
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

func descMessage(res ctlres.Resource) []string {
	if res.IsDeleting() {
		return []string{uiWaitMsgPrefix +
			ctlresm.NewDeleting(res).IsDoneApplying().Message}
	}
	return []string{}
}

func (c DeleteChange) isInoperableResource() bool {
	res := c.change.ExistingResource()
	for _, r := range inoperableResourceList {
		if r == (inoperableResourceRef{Name: res.Name(), GroupKind: res.GroupKind()}) {
			return true
		}
	}
	return false
}
