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
	res                  ctlres.Resource
	associatedRsFunc     func(ctlres.Resource, []ctlres.ResourceRef) ([]ctlres.Resource, error)
	specificResFactories []SpecificResFactory
}

type SpecificResFactory func(ctlres.Resource, []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef)

func NewConvergedResource(res ctlres.Resource,
	associatedRsFunc func(ctlres.Resource, []ctlres.ResourceRef) ([]ctlres.Resource, error),
	specificResFactories []SpecificResFactory) ConvergedResource {
	return ConvergedResource{res, associatedRsFunc, specificResFactories}
}

func (c ConvergedResource) IsDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	var descMsgs []string

	associatedRs, err := c.associatedRs()
	if err != nil {
		return ctlresm.DoneApplyState{}, nil, err
	}

	convergedResState, err := c.isResourceDoneApplying(c.res, associatedRs)
	if err != nil {
		return ctlresm.DoneApplyState{Done: true}, descMsgs, err
	}

	if convergedResState != nil {
		// If it is indeed done then take a quick way out and exit
		if convergedResState.Done {
			return *convergedResState, descMsgs, nil
		}
		if !convergedResState.Successful && len(associatedRs) > 0 {
			descMsgs = append(descMsgs, c.buildParentDescMsg(c.res, *convergedResState)...)
		}
	}

	// If resource explicitly opts out of associated resource waiting
	// exit quickly with parent resource state or success.
	// For example, CronJobs should be annotated with this to avoid
	// picking up failed Pods from previous runs.
	disableARWVal, disableARWFound := c.res.Annotations()[disableAssociatedResourcesWaitingAnnKey]
	if disableARWFound {
		if disableARWVal != "" {
			return ctlresm.DoneApplyState{Done: true}, descMsgs,
				fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
					disableAssociatedResourcesWaitingAnnKey, c.res.Description())
		} else {
			if convergedResState != nil {
				return *convergedResState, descMsgs, nil
			}
			return ctlresm.DoneApplyState{Done: true, Successful: true}, descMsgs, nil
		}
	}

	associatedRsStates := []ctlresm.DoneApplyState{}

	// Show associated resources even though we are waiting for the parent.
	// This additional info may be helpful in identifying what parent is waiting for.
	for _, res := range associatedRs {
		state, err := c.isResourceDoneApplying(res, associatedRs)
		if state == nil {
			state = &ctlresm.DoneApplyState{Done: true, Successful: true}
		}
		if err != nil {
			return *state, descMsgs, err
		}

		associatedRsStates = append(associatedRsStates, *state)
		descMsgs = append(descMsgs, c.buildChildDescMsg(res, *state)...)
	}

	// If parent state is present, ignore all associated resource states
	if convergedResState != nil {
		return *convergedResState, descMsgs, nil
	}

	for _, state := range associatedRsStates {
		if state.TerminallyFailed() {
			return state, descMsgs, nil
		}
	}

	for _, state := range associatedRsStates {
		if !state.Done {
			return state, descMsgs, nil
		}
	}

	return ctlresm.DoneApplyState{Done: true, Successful: true}, descMsgs, nil
}

func (c ConvergedResource) associatedRs() ([]ctlres.Resource, error) {
	if c.associatedRsFunc == nil {
		return nil, nil
	}
	for _, f := range c.specificResFactories {
		matchedRes, associatedResRefs := f(c.res, nil)
		// checking if interface is nil
		if !reflect.ValueOf(matchedRes).IsNil() {
			// Grab associated resources only if it's benefitial
			if len(associatedResRefs) > 0 {
				associatedRs, err := c.associatedRsFunc(c.res, associatedResRefs)
				if err != nil {
					return nil, err
				}
				return c.sortAssociatedRs(associatedRs), nil
			}
			break
		}
	}
	return nil, nil
}

func (c ConvergedResource) sortAssociatedRs(associatedRs []ctlres.Resource) []ctlres.Resource {
	convergedResKey := ctlres.NewUniqueResourceKey(c.res).String()

	// Sort so that resources show up more or less consistently
	sort.Slice(associatedRs, func(i, j int) bool {
		return associatedRs[i].Description() > associatedRs[j].Description()
	})

	// Remove possibly found parent resource
	for i, res := range associatedRs {
		if convergedResKey == ctlres.NewUniqueResourceKey(res).String() {
			associatedRs = append(associatedRs[:i], associatedRs[i+1:]...)
			break
		}
	}

	return associatedRs
}

func (c ConvergedResource) isResourceDoneApplying(res ctlres.Resource,
	associatedRs []ctlres.Resource) (*ctlresm.DoneApplyState, error) {

	for _, f := range c.specificResFactories {
		matchedRes, _ := f(res, associatedRs)
		// checking if interface is nil
		if !reflect.ValueOf(matchedRes).IsNil() {
			state := matchedRes.IsDoneApplying()
			return &state, nil
		}
	}
	return nil, nil
}

var (
	uiWaitChildPrefix    = color.New(color.Faint).Sprintf(" L ") // consistent with inspect tree view
	uiWaitMsgPrefix      = color.New(color.Faint).Sprintf(" ^ ")
	uiWaitChildMsgPrefix = "   " + uiWaitMsgPrefix
)

func (c ConvergedResource) buildParentDescMsg(res ctlres.Resource, state ctlresm.DoneApplyState) []string {
	if len(state.Message) > 0 {
		return []string{uiWaitMsgPrefix + state.Message}
	}
	return []string{}
}

func (c ConvergedResource) buildChildDescMsg(res ctlres.Resource, state ctlresm.DoneApplyState) []string {
	msgs := []string{fmt.Sprintf(uiWaitChildPrefix+"%s: waiting on %s", NewDoneApplyStateUI(state, nil).State, res.Description())}

	if len(state.Message) > 0 {
		msgs = append(msgs, uiWaitChildMsgPrefix+state.Message)
	}

	return msgs
}
