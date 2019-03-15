package clusterapply

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/fatih/color"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
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

var (
	uiWaitPrefix = color.New(color.Faint).Sprintf("%s", " L ") // consistent with inspect tree view
)

type AddOrUpdateChangeOpts struct {
	DefaultUpdateStrategy string
}

type AddOrUpdateChange struct {
	change              ctldiff.Change
	identifiedResources ctlres.IdentifiedResources
	changeFactory       ctldiff.ChangeFactory
	opts                AddOrUpdateChangeOpts
	ui                  UI
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

type SpecificResource interface {
	IsDoneApplying() ctlresm.DoneApplyState
}

func (c AddOrUpdateChange) IsDoneApplying() (ctlresm.DoneApplyState, error) {
	labeledResources := ctlres.NewLabeledResources(nil, c.identifiedResources)

	parentRes := c.change.NewResource()
	parentResKey := ctlres.NewUniqueResourceKey(parentRes).String()

	associatedRs, err := labeledResources.GetAssociated(parentRes)
	if err != nil {
		return ctlresm.DoneApplyState{}, err
	}

	sort.Slice(associatedRs, func(i, j int) bool {
		return associatedRs[i].Description() > associatedRs[j].Description()
	})

	states := []ctlresm.DoneApplyState{}
	allRs := append([]ctlres.Resource{parentRes}, associatedRs...)

	// Associated resources should also be checked
	// for example Deployment's Pods
	for _, res := range allRs {
		isParent := parentResKey == ctlres.NewUniqueResourceKey(res).String()

		if !isParent {
			c.ui.NotifyBegin(uiWaitPrefix+"waiting on %s", res.Description())
		}

		// TODO show all of them?
		state, err := c.isResourceDoneApplying(res, isParent)
		if err != nil {
			return state, err
		}

		if !isParent {
			msg := " ... "
			if state.Done {
				msg += "done"
			} else {
				msg += "in progress"
				if len(state.Message) > 0 {
					msg += ": " + state.Message
				}
			}
			c.ui.NotifyEnd(msg)
		}

		// TODO show parent status as well

		states = append(states, state)
	}

	// Find terminal unsuccessful state
	for _, state := range states {
		if state.Done && !state.Successful {
			return state, nil
		}
	}

	// Find first non-done case
	for _, state := range states {
		if !state.Done {
			return state, nil
		}
	}

	return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
}

func (c AddOrUpdateChange) isResourceDoneApplying(res ctlres.Resource, isParent bool) (ctlresm.DoneApplyState, error) {
	specificResFactories := []func(ctlres.Resource) SpecificResource{
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewDeleting(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewCRDvX(res) },
		// It's rare that Pod is directly created
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewPodvX(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewServicev1(res) },
	}

	for _, f := range specificResFactories {
		if !reflect.ValueOf(f(res)).IsNil() { // checking if interface is nil
			reloadedResource, err := c.identifiedResources.Get(res)
			if err != nil {
				// Non-parent resource may go away, and that can be ignored
				if !isParent && errors.IsNotFound(err) {
					return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
				}
				return ctlresm.DoneApplyState{}, err
			}

			return f(reloadedResource).IsDoneApplying(), nil
		}
	}

	return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
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
