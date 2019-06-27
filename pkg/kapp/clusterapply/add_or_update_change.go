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

	disableAssociatedResourcesWaitingAnnKey = "kapp.k14s.io/disable-associated-resources-wait" // valid value is ''
)

var (
	uiWaitPrefix       = color.New(color.Faint).Sprintf(" L ") // consistent with inspect tree view
	uiWaitParentPrefix = color.New(color.Faint).Sprintf(" ^ ")
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
	parentRes, associatedRs, err := c.findParentAndAssociatedRes()
	if err != nil {
		return ctlresm.DoneApplyState{Done: true}, err
	}

	parentResState, err := c.isResourceDoneApplying(parentRes, true)
	if err != nil {
		return ctlresm.DoneApplyState{Done: true}, err
	}

	if parentResState != nil {
		// If it is indeed done then take a quick way out and exit
		if parentResState.Done {
			return *parentResState, nil
		}

		if !parentResState.Successful && len(associatedRs) > 0 {
			c.notify(parentRes, true, *parentResState)
		}
	}

	// If resource explicitly opts out of associated resource waiting
	// exit quickly with parent resource state or success.
	// For example, CronJobs should be annotated with this to avoid
	// picking up failed Pods from previous runs.
	disableARWVal, disableARWFound := parentRes.Annotations()[disableAssociatedResourcesWaitingAnnKey]
	if disableARWFound {
		if disableARWVal != "" {
			return ctlresm.DoneApplyState{Done: true},
				fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
					disableAssociatedResourcesWaitingAnnKey, parentRes.Description())
		} else {
			if parentResState != nil {
				return *parentResState, nil
			}
			return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
		}
	}

	associatedRsStates := []ctlresm.DoneApplyState{}

	// Show associated resources even though we are waiting for the parent.
	// This additional info may be helpful in identifying what parent is waiting for.
	for _, res := range associatedRs {
		state, err := c.isResourceDoneApplying(res, false)
		if state == nil {
			state = &ctlresm.DoneApplyState{Done: true, Successful: true}
		}
		if err != nil {
			return *state, err
		}

		associatedRsStates = append(associatedRsStates, *state)
		c.notify(res, false, *state)
	}

	// If parent state is present, ignore all associated resource states
	if parentResState != nil {
		return *parentResState, nil
	}

	for _, state := range associatedRsStates {
		if state.TerminallyFailed() {
			return state, nil
		}
	}

	for _, state := range associatedRsStates {
		if !state.Done {
			return state, nil
		}
	}

	return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
}

func (c AddOrUpdateChange) findParentAndAssociatedRes() (ctlres.Resource, []ctlres.Resource, error) {
	labeledResources := ctlres.NewLabeledResources(nil, c.identifiedResources)

	parentRes := c.change.NewResource()
	parentResKey := ctlres.NewUniqueResourceKey(parentRes).String()

	associatedRs, err := labeledResources.GetAssociated(parentRes)
	if err != nil {
		return nil, nil, err
	}

	// Sort so that resources show up more or less consistently
	sort.Slice(associatedRs, func(i, j int) bool {
		return associatedRs[i].Description() > associatedRs[j].Description()
	})

	// Remove possibly found parent resource
	for i, res := range associatedRs {
		if parentResKey == ctlres.NewUniqueResourceKey(res).String() {
			associatedRs = append(associatedRs[:i], associatedRs[i+1:]...)
			break
		}
	}

	return parentRes, associatedRs, nil
}

func (c AddOrUpdateChange) isResourceDoneApplying(res ctlres.Resource, isParent bool) (*ctlresm.DoneApplyState, error) {
	specificResFactories := []func(ctlres.Resource) SpecificResource{
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewDeleting(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewCRDvX(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewCorev1Pod(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewCorev1Service(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewAppsv1Deployment(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewAppsv1DaemonSet(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewBatchv1Job(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewBatchvxCronJob(res) },
	}

	for _, f := range specificResFactories {
		if !reflect.ValueOf(f(res)).IsNil() { // checking if interface is nil
			reloadedResource, err := c.identifiedResources.Get(res)
			if err != nil {
				// Non-parent resource may go away, and that can be ignored
				if !isParent && errors.IsNotFound(err) {
					return &ctlresm.DoneApplyState{Done: true, Successful: true}, nil
				}
				return nil, err
			}

			state := f(reloadedResource).IsDoneApplying()
			return &state, nil
		}
	}

	return nil, nil
}

func (c AddOrUpdateChange) notify(res ctlres.Resource, isParent bool, state ctlresm.DoneApplyState) {
	msg := fmt.Sprintf(uiWaitPrefix+"waiting on %s", res.Description())
	if isParent {
		msg = uiWaitParentPrefix
	}

	// End of notification
	msg += " ... "
	if state.Done {
		msg += "done"
	} else {
		msg += "in progress"
		if len(state.Message) > 0 {
			msg += ": " + state.Message
		}
	}

	c.ui.Notify(msg)
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
