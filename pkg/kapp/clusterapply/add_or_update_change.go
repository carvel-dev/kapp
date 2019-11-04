package clusterapply

import (
	"fmt"
	"time"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"github.com/k14s/kapp/pkg/kapp/util"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
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
		createdRes, err := c.identifiedResources.Create(c.change.NewResource())
		if err != nil {
			return err
		}

		err = c.recordAppliedResource(createdRes)
		if err != nil {
			return err
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
					return c.tryToResolveConflict(err)
				}
				return err
			}

			err = c.recordAppliedResource(updatedRes)
			if err != nil {
				return err
			}

		case updateStrategyFallbackOnReplaceAnnValue:
			updatedRes, err := c.identifiedResources.Update(newRes)
			if err != nil {
				if errors.IsInvalid(err) {
					return c.replace()
				}
				return err
			}

			err = c.recordAppliedResource(updatedRes)
			if err != nil {
				return err
			}

		case updateStrategyAlwaysReplaceAnnValue:
			return c.replace()

		default:
			return fmt.Errorf("Unknown update strategy: %s", strategy)
		}
	}

	return nil
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

func (a AddOrUpdateChange) tryToResolveConflict(origErr error) error {
	errMsgPrefix := "Failed to update due to resource conflict "

	for i := 0; i < 10; i++ {
		latestExistingRes, err := a.identifiedResources.Get(a.change.ExistingResource())
		if err != nil {
			return err
		}

		changeSet := a.changeSetFactory.New([]ctlres.Resource{latestExistingRes}, []ctlres.Resource{a.change.AppliedResource()})

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
			return err
		}

		return a.recordAppliedResource(updatedRes)
	}

	return fmt.Errorf(errMsgPrefix+"(tried multiple times): %s", origErr)
}

type SpecificResource interface {
	IsDoneApplying() ctlresm.DoneApplyState
}

func (c AddOrUpdateChange) IsDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	labeledResources := ctlres.NewLabeledResources(nil, c.identifiedResources, logger.NewTODOLogger())

	// Refresh resource with latest changes from the server
	parentRes, err := c.identifiedResources.Get(c.change.NewResource())
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	associatedRs, err := labeledResources.GetAssociated(parentRes)
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	return NewConvergedResource(parentRes, associatedRs).IsDoneApplying()
}

func (c AddOrUpdateChange) recordAppliedResource(savedRes ctlres.Resource) error {
	reloadedSavedRes := savedRes // first time, try using memory copy

	return util.Retry(time.Second, time.Minute, func() (bool, error) {
		// subsequent times try to retrieve latest copy,
		// for example, ServiceAccount seems to change immediately
		if reloadedSavedRes == nil {
			res, err := c.identifiedResources.Get(savedRes)
			if err != nil {
				return false, err
			}

			reloadedSavedRes = res
		}

		savedResWithHistory := c.changeFactory.NewResourceWithHistory(reloadedSavedRes)

		resWithHistory, err := savedResWithHistory.RecordLastAppliedResource(c.change.AppliedResource())
		if err != nil {
			return true, fmt.Errorf("Recording last applied resource: %s", err)
		}

		_, err = c.identifiedResources.Update(resWithHistory)
		if err != nil {
			reloadedSavedRes = nil // Get again
			return false, fmt.Errorf("Saving record of last applied resource: %s", err)
		}

		return true, nil
	})
}
