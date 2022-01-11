// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"time"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	createStrategyAnnKey                                                = "kapp.k14s.io/create-strategy"
	createStrategyPlainAnnValue            ClusterChangeApplyStrategyOp = ""
	createStrategyFallbackOnUpdateAnnValue ClusterChangeApplyStrategyOp = "fallback-on-update"

	updateStrategyAnnKey                                                 = "kapp.k14s.io/update-strategy"
	updateStrategyPlainAnnValue             ClusterChangeApplyStrategyOp = ""
	updateStrategyFallbackOnReplaceAnnValue ClusterChangeApplyStrategyOp = "fallback-on-replace"
	updateStrategyAlwaysReplaceAnnValue     ClusterChangeApplyStrategyOp = "always-replace"
	updateStrategySkipAnnValue              ClusterChangeApplyStrategyOp = "skip"
)

type AddOrUpdateChangeOpts struct {
	DefaultUpdateStrategy   string
	ServerSideApply         bool
	ServerSideForceConflict bool
	FieldManagerName        string
}

type AddOrUpdateChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
	changeFactory       ctldiff.ChangeFactory
	changeSetFactory    ctldiff.ChangeSetFactory
	opts                AddOrUpdateChangeOpts
}

func (c AddOrUpdateChange) ApplyStrategy() (ApplyStrategy, error) {
	op := c.change.Op()

	switch op {
	case ctldiff.ChangeOpAdd:
		newRes := c.change.NewResource()

		strategy, found := newRes.Annotations()[createStrategyAnnKey]
		if !found {
			strategy = string(createStrategyPlainAnnValue)
		}

		switch ClusterChangeApplyStrategyOp(strategy) {
		case createStrategyPlainAnnValue:
			return AddPlainStrategy{newRes, c}, nil

		case createStrategyFallbackOnUpdateAnnValue:
			return AddOrFallbackOnUpdateStrategy{newRes, c}, nil

		default:
			return nil, fmt.Errorf("Unknown create strategy: %s", strategy)
		}

	case ctldiff.ChangeOpUpdate:
		newRes := c.change.NewResource()

		strategy, found := newRes.Annotations()[updateStrategyAnnKey]
		if !found {
			strategy = c.opts.DefaultUpdateStrategy
		}

		switch ClusterChangeApplyStrategyOp(strategy) {
		case updateStrategyPlainAnnValue:
			return UpdatePlainStrategy{newRes, c}, nil

		case updateStrategyFallbackOnReplaceAnnValue:
			return UpdateOrFallbackOnReplaceStrategy{newRes, c}, nil

		case updateStrategyAlwaysReplaceAnnValue:
			return UpdateAlwaysReplaceStrategy{c}, nil

		case updateStrategySkipAnnValue:
			return UpdateSkipStrategy{c}, nil

		default:
			return nil, fmt.Errorf("Unknown update strategy: %s", strategy)
		}

	default:
		return nil, fmt.Errorf("Unknown add-or-update op: %s", op)
	}
}

func (c AddOrUpdateChange) replace() error {
	// TODO do we have to wait for delete to finish?
	err := c.identifiedResources.Delete(c.change.ExistingResource())
	if err != nil {
		return err
	}

	// Wait for the resource to be fully deleted
	for {
		_, exists, err := c.identifiedResources.Exists(c.change.ExistingResource(), ctlres.ExistsOpts{})
		if err != nil {
			return err
		}
		if !exists {
			break
		}
		time.Sleep(1 * time.Second)
	}

	updatedRes, err := c.identifiedResources.Create(c.change.AppliedResource())
	if err != nil {
		return err
	}

	return c.recordAppliedResource(updatedRes)
}

func (c AddOrUpdateChange) tryToResolveUpdateConflict(
	origErr error, updateFallbackFunc func(error) error) error {

	errMsgPrefix := "Failed to update due to resource conflict "

	for i := 0; i < 10; i++ {
		latestExistingRes, err := c.identifiedResources.Get(c.change.ExistingResource())
		if err != nil {
			return err
		}

		changeSet := c.changeSetFactory.New([]ctlres.Resource{latestExistingRes},
			[]ctlres.Resource{c.change.AppliedResource()})

		recalcChanges, err := changeSet.Calculate(context.TODO())
		if err != nil {
			return err
		}

		if len(recalcChanges) != 1 {
			return fmt.Errorf("Expected exactly one change when recalculating conflicting change")
		}
		if recalcChanges[0].Op() != ctldiff.ChangeOpUpdate {
			return fmt.Errorf("Expected recalculated change to be an update")
		}
		if recalcChanges[0].OpsDiff().MinimalMD5() != c.change.OpsDiff().MinimalMD5() {
			return fmt.Errorf(errMsgPrefix+"(approved diff no longer matches): %s", origErr)
		}

		updatedRes, err := c.identifiedResources.Update(recalcChanges[0].NewResource())
		if err != nil {
			if errors.IsConflict(err) {
				continue
			}
			return updateFallbackFunc(err)
		}

		return c.recordAppliedResource(updatedRes)
	}

	return fmt.Errorf(errMsgPrefix+"(tried multiple times): %s", origErr)
}

func (c AddOrUpdateChange) tryToUpdateAfterCreateConflict() error {
	var lastUpdateErr error

	for i := 0; i < 10; i++ {
		latestExistingRes, err := c.identifiedResources.Get(c.change.NewResource())
		if err != nil {
			return err
		}

		changeSet := c.changeSetFactory.New([]ctlres.Resource{latestExistingRes},
			[]ctlres.Resource{c.change.AppliedResource()})

		recalcChanges, err := changeSet.Calculate(context.TODO())
		if err != nil {
			return err
		}

		if len(recalcChanges) != 1 {
			return fmt.Errorf("Expected exactly one change when recalculating conflicting change")
		}
		if recalcChanges[0].Op() != ctldiff.ChangeOpUpdate {
			return fmt.Errorf("Expected recalculated change to be an update")
		}

		updatedRes, err := c.identifiedResources.Update(recalcChanges[0].NewResource())
		if err != nil {
			if errors.IsConflict(err) {
				lastUpdateErr = err
				continue
			}
			return err
		}

		return c.recordAppliedResource(updatedRes)
	}

	return fmt.Errorf("Failed to update (after trying to create) "+
		"due to resource conflict (tried multiple times): %s", lastUpdateErr)
}

func (c AddOrUpdateChange) recordAppliedResource(savedRes ctlres.Resource) error {
	savedResWithHistory := c.changeFactory.NewResourceWithHistory(savedRes)

	// It may not be benefitial to record last applied conf
	// onto resource. This could be useful for resources that
	// are very large, hence go over annotation value max length.
	if !savedResWithHistory.AllowsRecordingLastApplied() {
		return nil
	}

	// Calculate change _once_ against what was returned from the server
	// (ie changes applied by the webhooks on the server, etc _but
	// not by other controllers)
	applyChange, err := savedResWithHistory.CalculateChange(c.change.AppliedResource())
	if err != nil {
		return fmt.Errorf("Calculating change after the save: %s", err)
	}

	// Record last applied change on the latest version of a resource
	recordHistoryPatch, err := savedResWithHistory.RecordLastAppliedResource(applyChange)
	if err != nil {
		return fmt.Errorf("Recording last applied resource: %s", err)
	}

	_, err = c.identifiedResources.Patch(savedRes, types.MergePatchType, recordHistoryPatch, ctlres.PatchOpts{DryRun: false})
	return err
}

type AddPlainStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c AddPlainStrategy) Op() ClusterChangeApplyStrategyOp { return createStrategyPlainAnnValue }

func (c AddPlainStrategy) Apply() error {
	createdRes, err := c.aou.identifiedResources.Create(c.newRes)
	if err != nil {
		return err
	}

	// Create is recorded in the metadata.fieldManagers as
	// Update operation creating distinct field manager from Apply operation,
	// which means that these fields wont be updateable using SSA.
	// To fix it, we change operation to be "Apply"
	// See https://github.com/kubernetes/kubernetes/issues/107417 for details
	if c.aou.opts.ServerSideApply {
		createdRes, err = c.aou.identifiedResources.Patch(createdRes, types.JSONPatchType, []byte(`
[
	{ "op": "test", "path": "/metadata/managedFields/0/manager", "value": "`+c.aou.opts.FieldManagerName+`" },
	{ "op": "replace", "path": "/metadata/managedFields/0/operation", "value": "Apply" }
]
`), ctlres.PatchOpts{DryRun: false})
		if err != nil {
			// TODO: potentially patch can fail if '"op": "test"' fails, which can happen if another
			// controller changes managedFields. We
			return err
		}

	}

	return c.aou.recordAppliedResource(createdRes)
}

type AddOrFallbackOnUpdateStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c AddOrFallbackOnUpdateStrategy) Op() ClusterChangeApplyStrategyOp {
	return createStrategyFallbackOnUpdateAnnValue
}

func (c AddOrFallbackOnUpdateStrategy) Apply() (err error) {
	var createdRes ctlres.Resource
	if c.aou.opts.ServerSideApply {
		resBytes, err := c.newRes.AsYAMLBytes()
		if err != nil {
			return err
		}

		// Apply patch is like upsert, combining create + update, no need to fallback on error
		createdRes, err = c.aou.identifiedResources.Patch(c.newRes, types.ApplyPatchType, resBytes, ctlres.PatchOpts{DryRun: false})
		if err != nil {
			return err
		}
	} else {
		createdRes, err = c.aou.identifiedResources.Create(c.newRes)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return c.aou.tryToUpdateAfterCreateConflict()
			}
			return err
		}
	}

	return c.aou.recordAppliedResource(createdRes)
}

type UpdatePlainStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c UpdatePlainStrategy) Op() ClusterChangeApplyStrategyOp { return updateStrategyPlainAnnValue }

func (c UpdatePlainStrategy) Apply() error {
	var updatedRes ctlres.Resource
	var err error

	if c.aou.opts.ServerSideApply {
		updatedRes, err = ctlres.WithIdentityAnnotation(c.newRes, func(r ctlres.Resource) (ctlres.Resource, error) {
			resBytes, _ := r.AsYAMLBytes()
			return c.aou.identifiedResources.Patch(r, types.ApplyPatchType, resBytes, ctlres.PatchOpts{DryRun: false})
		})
	} else {
		updatedRes, err = c.aou.identifiedResources.Update(c.newRes)
	}
	if err != nil {
		if errors.IsConflict(err) {
			return c.aou.tryToResolveUpdateConflict(err, func(err error) error { return err })
		}
		return err
	}

	return c.aou.recordAppliedResource(updatedRes)
}

type UpdateOrFallbackOnReplaceStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c UpdateOrFallbackOnReplaceStrategy) Op() ClusterChangeApplyStrategyOp {
	return updateStrategyFallbackOnReplaceAnnValue
}

func (c UpdateOrFallbackOnReplaceStrategy) Apply() error {
	replaceIfIsInvalidErrFunc := func(err error) error {
		if errors.IsInvalid(err) {
			return c.aou.replace()
		}
		return err
	}

	var updatedRes ctlres.Resource
	var err error

	if c.aou.opts.ServerSideApply {
		updatedRes, err = ctlres.WithIdentityAnnotation(c.newRes, func(r ctlres.Resource) (ctlres.Resource, error) {
			resBytes, _ := r.AsYAMLBytes()
			return c.aou.identifiedResources.Patch(r, types.ApplyPatchType, resBytes, ctlres.PatchOpts{DryRun: false})
		})
	} else {
		updatedRes, err = c.aou.identifiedResources.Update(c.newRes)
	}

	if err != nil {
		//TODO: find out if SSA conflicts worth retrying
		if errors.IsConflict(err) {
			return c.aou.tryToResolveUpdateConflict(err, replaceIfIsInvalidErrFunc)
		}
		return replaceIfIsInvalidErrFunc(err)
	}

	return c.aou.recordAppliedResource(updatedRes)
}

type UpdateAlwaysReplaceStrategy struct {
	aou AddOrUpdateChange
}

func (c UpdateAlwaysReplaceStrategy) Op() ClusterChangeApplyStrategyOp {
	return updateStrategyAlwaysReplaceAnnValue
}

func (c UpdateAlwaysReplaceStrategy) Apply() error {
	return c.aou.replace()
}

type UpdateSkipStrategy struct {
	aou AddOrUpdateChange
}

func (c UpdateSkipStrategy) Op() ClusterChangeApplyStrategyOp {
	return updateStrategySkipAnnValue
}

func (c UpdateSkipStrategy) Apply() error { return nil }
