// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"
	"time"

	ctlconf "carvel.dev/kapp/pkg/kapp/config"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"carvel.dev/kapp/pkg/kapp/util"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	createStrategyAnnKey                                                      = "kapp.k14s.io/create-strategy"
	createStrategyPlainAnnValue                  ClusterChangeApplyStrategyOp = ""
	createStrategyFallbackOnUpdateAnnValue       ClusterChangeApplyStrategyOp = "fallback-on-update"
	createStrategyFallbackOnUpdateOrNoopAnnValue ClusterChangeApplyStrategyOp = "fallback-on-update-or-noop"

	updateStrategyAnnKey                                                 = "kapp.k14s.io/update-strategy"
	updateStrategyPlainAnnValue             ClusterChangeApplyStrategyOp = ""
	updateStrategyFallbackOnReplaceAnnValue ClusterChangeApplyStrategyOp = "fallback-on-replace"
	updateStrategyAlwaysReplaceAnnValue     ClusterChangeApplyStrategyOp = "always-replace"
	updateStrategySkipAnnValue              ClusterChangeApplyStrategyOp = "skip"
)

type AddOrUpdateChangeOpts struct {
	DefaultUpdateStrategy string
}

type AddOrUpdateChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
	changeFactory       ctldiff.ChangeFactory
	changeSetFactory    ctldiff.ChangeSetFactory
	opts                AddOrUpdateChangeOpts
	diffMaskRules       []ctlconf.DiffMaskRule
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

		case createStrategyFallbackOnUpdateOrNoopAnnValue:
			return AddOrFallbackOnUpdateOrNoopStrategy{newRes, c}, nil

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

		recalcChanges, err := changeSet.Calculate()
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
			errMsg := fmt.Sprintf(errMsgPrefix+"(approved diff no longer matches): %s", origErr)

			textDiff, err := recalcChanges[0].ConfigurableTextDiff().Masked(c.diffMaskRules)
			if err != nil {
				return fmt.Errorf(errMsg)
			}
			return fmt.Errorf("%s: Recalculated diff:\n%s", errMsg, textDiff.MinimalString())
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

	return fmt.Errorf(errMsgPrefix+"(tried multiple times): %w", origErr)
}

func (c AddOrUpdateChange) tryToUpdateAfterCreateConflict(allowNoopUpdates bool) error {
	var lastUpdateErr error

	for i := 0; i < 10; i++ {
		latestExistingRes, err := c.identifiedResources.Get(c.change.NewResource())
		if err != nil {
			return err
		}

		changeSet := c.changeSetFactory.New([]ctlres.Resource{latestExistingRes},
			[]ctlres.Resource{c.change.AppliedResource()})

		recalcChanges, err := changeSet.Calculate()
		if err != nil {
			return err
		}

		if len(recalcChanges) != 1 {
			return fmt.Errorf("Expected exactly one change when recalculating conflicting change")
		}
		switch recalcChanges[0].Op() {
		case ctldiff.ChangeOpNoop:
			if allowNoopUpdates {
				return nil
			}
			return fmt.Errorf("Recalculated change was a noop but those are not allowed (try adding annotation kapp.k14s.io/create-strategy: fallback-on-update-or-noop)")
		case ctldiff.ChangeOpUpdate:
			updatedRes, err := c.identifiedResources.Update(recalcChanges[0].NewResource())
			if err != nil {
				if errors.IsConflict(err) {
					lastUpdateErr = err
					continue
				}
				return err
			}
			return c.recordAppliedResource(updatedRes)
		default:
			return fmt.Errorf("Expected recalculated change to be an update or noop")
		}
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
		return fmt.Errorf("Calculating change after the save: %w", err)
	}

	// first time, try using memory copy
	latestResWithHistory := &savedResWithHistory

	return util.Retry(time.Second, time.Minute, func() (bool, error) {
		// subsequent times try to retrieve latest copy,
		// for example, ServiceAccount seems to change immediately
		if latestResWithHistory == nil {
			res, err := c.identifiedResources.Get(savedRes)
			if err != nil {
				return false, err
			}

			resWithHistory := c.changeFactory.NewResourceWithHistory(res)
			latestResWithHistory = &resWithHistory
		}

		// Record last applied change on the latest version of a resource
		latestResWithHistoryUpdated, madeAnyModifications, err := latestResWithHistory.RecordLastAppliedResource(applyChange)
		if err != nil {
			return true, fmt.Errorf("Recording last applied resource: %w", err)
		}

		// when annotation value max length exceed then don't record the resource hence not making any modification
		if !madeAnyModifications {
			return true, nil
		}

		_, err = c.identifiedResources.Update(latestResWithHistoryUpdated)
		if err != nil {
			latestResWithHistory = nil // Get again
			return false, fmt.Errorf("Saving record of last applied resource: %w", err)
		}

		return true, nil
	})
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

	return c.aou.recordAppliedResource(createdRes)
}

type AddOrFallbackOnUpdateStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c AddOrFallbackOnUpdateStrategy) Op() ClusterChangeApplyStrategyOp {
	return createStrategyFallbackOnUpdateAnnValue
}

func (c AddOrFallbackOnUpdateStrategy) Apply() error {
	createdRes, err := c.aou.identifiedResources.Create(c.newRes)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return c.aou.tryToUpdateAfterCreateConflict(false)
		}
		return err
	}

	return c.aou.recordAppliedResource(createdRes)
}

type AddOrFallbackOnUpdateOrNoopStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c AddOrFallbackOnUpdateOrNoopStrategy) Op() ClusterChangeApplyStrategyOp {
	return createStrategyFallbackOnUpdateOrNoopAnnValue
}

func (c AddOrFallbackOnUpdateOrNoopStrategy) Apply() error {
	createdRes, err := c.aou.identifiedResources.Create(c.newRes)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return c.aou.tryToUpdateAfterCreateConflict(true)
		}
		return err
	}

	return c.aou.recordAppliedResource(createdRes)
}

type UpdatePlainStrategy struct {
	newRes ctlres.Resource
	aou    AddOrUpdateChange
}

func (c UpdatePlainStrategy) Op() ClusterChangeApplyStrategyOp { return updateStrategyPlainAnnValue }

func (c UpdatePlainStrategy) Apply() error {
	updatedRes, err := c.aou.identifiedResources.Update(c.newRes)
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

	updatedRes, err := c.aou.identifiedResources.Update(c.newRes)
	if err != nil {
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
