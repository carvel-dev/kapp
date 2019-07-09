package clusterapply

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/fatih/color"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

const (
	disableAssociatedResourcesWaitingAnnKey = "kapp.k14s.io/disable-associated-resources-wait" // valid value is ''
)

type ConvergedResource struct {
	res          ctlres.Resource
	associatedRs []ctlres.Resource
}

func NewConvergedResource(res ctlres.Resource, associatedRs []ctlres.Resource) ConvergedResource {
	return ConvergedResource{res, associatedRs}
}

func (c ConvergedResource) IsDoneApplying(ui UI) (ctlresm.DoneApplyState, error) {
	convergedRes, associatedRs, err := c.findParentAndAssociatedRes()
	if err != nil {
		return ctlresm.DoneApplyState{Done: true}, err
	}

	convergedResState, err := c.isResourceDoneApplying(convergedRes)
	if err != nil {
		return ctlresm.DoneApplyState{Done: true}, err
	}

	if convergedResState != nil {
		// If it is indeed done then take a quick way out and exit
		if convergedResState.Done {
			return *convergedResState, nil
		}

		if !convergedResState.Successful && len(associatedRs) > 0 {
			c.notify(ui, convergedRes, true, *convergedResState)
		}
	}

	// If resource explicitly opts out of associated resource waiting
	// exit quickly with parent resource state or success.
	// For example, CronJobs should be annotated with this to avoid
	// picking up failed Pods from previous runs.
	disableARWVal, disableARWFound := convergedRes.Annotations()[disableAssociatedResourcesWaitingAnnKey]
	if disableARWFound {
		if disableARWVal != "" {
			return ctlresm.DoneApplyState{Done: true},
				fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
					disableAssociatedResourcesWaitingAnnKey, convergedRes.Description())
		} else {
			if convergedResState != nil {
				return *convergedResState, nil
			}
			return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
		}
	}

	associatedRsStates := []ctlresm.DoneApplyState{}

	// Show associated resources even though we are waiting for the parent.
	// This additional info may be helpful in identifying what parent is waiting for.
	for _, res := range associatedRs {
		state, err := c.isResourceDoneApplying(res)
		if state == nil {
			state = &ctlresm.DoneApplyState{Done: true, Successful: true}
		}
		if err != nil {
			return *state, err
		}

		associatedRsStates = append(associatedRsStates, *state)
		c.notify(ui, res, false, *state)
	}

	// If parent state is present, ignore all associated resource states
	if convergedResState != nil {
		return *convergedResState, nil
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

func (c ConvergedResource) findParentAndAssociatedRes() (ctlres.Resource, []ctlres.Resource, error) {
	convergedRes := c.res
	convergedResKey := ctlres.NewUniqueResourceKey(convergedRes).String()

	// Sort so that resources show up more or less consistently
	sort.Slice(c.associatedRs, func(i, j int) bool {
		return c.associatedRs[i].Description() > c.associatedRs[j].Description()
	})

	// Remove possibly found parent resource
	for i, res := range c.associatedRs {
		if convergedResKey == ctlres.NewUniqueResourceKey(res).String() {
			c.associatedRs = append(c.associatedRs[:i], c.associatedRs[i+1:]...)
			break
		}
	}

	return convergedRes, c.associatedRs, nil
}

func (c ConvergedResource) isResourceDoneApplying(res ctlres.Resource) (*ctlresm.DoneApplyState, error) {
	specificResFactories := []func(ctlres.Resource) SpecificResource{
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewDeleting(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewApiExtensionsVxCRD(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewCoreV1Pod(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewCoreV1Service(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewAppsV1Deployment(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewAppsV1DaemonSet(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewBatchV1Job(res) },
		func(res ctlres.Resource) SpecificResource { return ctlresm.NewBatchVxCronJob(res) },
	}

	for _, f := range specificResFactories {
		if !reflect.ValueOf(f(res)).IsNil() { // checking if interface is nil
			state := f(res).IsDoneApplying()
			return &state, nil
		}
	}

	return nil, nil
}

var (
	uiWaitPrefix       = color.New(color.Faint).Sprintf(" L ") // consistent with inspect tree view
	uiWaitParentPrefix = color.New(color.Faint).Sprintf(" ^ ")
)

func (c ConvergedResource) notify(ui UI, res ctlres.Resource, isParent bool, state ctlresm.DoneApplyState) {
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
	}
	if len(state.Message) > 0 {
		msg += ": " + state.Message
	}

	ui.Notify(msg)
}
