package clusterapply

import (
	"fmt"
	"time"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/k14s/kapp/pkg/kapp/util"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	createStrategyAnnKey                   = "kapp.k14s.io/create-strategy"
	createStrategyCreateAnnValue           = ""
	createStrategyFallbackOnUpdateAnnValue = "fallback-on-update"

	updateStrategyAnnKey                    = "kapp.k14s.io/update-strategy"
	updateStrategyUpdateAnnValue            = ""
	updateStrategyFallbackOnReplaceAnnValue = "fallback-on-replace"
	updateStrategyAlwaysReplaceAnnValue     = "always-replace"
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
}

func (c AddOrUpdateChange) Apply() error {
	op := c.change.Op()

	switch op {
	case ctldiff.ChangeOpAdd:
		newRes := c.change.NewResource()

		strategy, found := newRes.Annotations()[createStrategyAnnKey]
		if !found {
			strategy = createStrategyCreateAnnValue
		}

		switch strategy {
		case createStrategyCreateAnnValue:
			createdRes, err := c.identifiedResources.Create(newRes)
			if err != nil {
				return err
			}

			return c.recordAppliedResource(createdRes)

		case createStrategyFallbackOnUpdateAnnValue:
			createdRes, err := c.identifiedResources.Create(newRes)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					return c.tryToUpdateAfterCreateConflict()
				}
				return err
			}

			return c.recordAppliedResource(createdRes)

		default:
			return fmt.Errorf("Unknown create strategy: %s", strategy)
		}

	case ctldiff.ChangeOpUpdate:
		newRes := c.change.NewResource()

		strategy, found := newRes.Annotations()[updateStrategyAnnKey]
		if !found {
			strategy = c.opts.DefaultUpdateStrategy
		}

		switch strategy {
		case updateStrategyUpdateAnnValue:
			updatedRes, err := c.identifiedResources.Update(newRes)
			if err != nil {
				if errors.IsConflict(err) {
					return c.tryToResolveUpdateConflict(err, func(err error) error { return err })
				}
				return err
			}

			return c.recordAppliedResource(updatedRes)

		case updateStrategyFallbackOnReplaceAnnValue:
			replaceIfIsInvalidErrFunc := func(err error) error {
				if errors.IsInvalid(err) {
					return c.replace()
				}
				return err
			}

			updatedRes, err := c.identifiedResources.Update(newRes)
			if err != nil {
				if errors.IsConflict(err) {
					return c.tryToResolveUpdateConflict(err, replaceIfIsInvalidErrFunc)
				}
				return replaceIfIsInvalidErrFunc(err)
			}

			return c.recordAppliedResource(updatedRes)

		case updateStrategyAlwaysReplaceAnnValue:
			return c.replace()

		default:
			return fmt.Errorf("Unknown update strategy: %s", strategy)
		}

	default:
		return fmt.Errorf("Unknown add-or-update op: %s", op)
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
		exists, err := c.identifiedResources.Exists(c.change.ExistingResource())
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

func (a AddOrUpdateChange) tryToResolveUpdateConflict(
	origErr error, updateFallbackFunc func(error) error) error {

	errMsgPrefix := "Failed to update due to resource conflict "

	for i := 0; i < 10; i++ {
		latestExistingRes, err := a.identifiedResources.Get(a.change.ExistingResource())
		if err != nil {
			return err
		}

		changeSet := a.changeSetFactory.New([]ctlres.Resource{latestExistingRes},
			[]ctlres.Resource{a.change.AppliedResource()})

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
		if recalcChanges[0].OpsDiff().MinimalMD5() != a.change.OpsDiff().MinimalMD5() {
			return fmt.Errorf(errMsgPrefix+"(approved diff no longer matches): %s", origErr)
		}

		updatedRes, err := a.identifiedResources.Update(recalcChanges[0].NewResource())
		if err != nil {
			if errors.IsConflict(err) {
				continue
			}
			return updateFallbackFunc(err)
		}

		return a.recordAppliedResource(updatedRes)
	}

	return fmt.Errorf(errMsgPrefix+"(tried multiple times): %s", origErr)
}

func (a AddOrUpdateChange) tryToUpdateAfterCreateConflict() error {
	var lastUpdateErr error

	for i := 0; i < 10; i++ {
		latestExistingRes, err := a.identifiedResources.Get(a.change.NewResource())
		if err != nil {
			return err
		}

		changeSet := a.changeSetFactory.New([]ctlres.Resource{latestExistingRes},
			[]ctlres.Resource{a.change.AppliedResource()})

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

		updatedRes, err := a.identifiedResources.Update(recalcChanges[0].NewResource())
		if err != nil {
			if errors.IsConflict(err) {
				lastUpdateErr = err
				continue
			}
			return err
		}

		return a.recordAppliedResource(updatedRes)
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
		return err
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
		latestResWithHistoryUpdated, err := latestResWithHistory.RecordLastAppliedResource(applyChange)
		if err != nil {
			return true, fmt.Errorf("Recording last applied resource: %s", err)
		}

		_, err = c.identifiedResources.Update(latestResWithHistoryUpdated)
		if err != nil {
			latestResWithHistory = nil // Get again
			return false, fmt.Errorf("Saving record of last applied resource: %s", err)
		}

		return true, nil
	})
}
