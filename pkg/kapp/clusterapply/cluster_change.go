// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"
	"strings"

	ctlconf "carvel.dev/kapp/pkg/kapp/config"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
	uierrs "github.com/cppforlife/go-cli-ui/errors"
)

const (
	disableWaitAnnKey = "kapp.k14s.io/disable-wait" // valid values: ''
)

type ClusterChangeApplyOp string

const (
	ClusterChangeApplyOpAdd    ClusterChangeApplyOp = "add"
	ClusterChangeApplyOpDelete ClusterChangeApplyOp = "delete"
	ClusterChangeApplyOpUpdate ClusterChangeApplyOp = "update"
	ClusterChangeApplyOpNoop   ClusterChangeApplyOp = "noop"
	ClusterChangeApplyOpExists ClusterChangeApplyOp = "exists"
)

type ClusterChangeWaitOp string

const (
	ClusterChangeWaitOpOK     ClusterChangeWaitOp = "ok"
	ClusterChangeWaitOpDelete ClusterChangeWaitOp = "delete"
	ClusterChangeWaitOpNoop   ClusterChangeWaitOp = "noop"
)

type ClusterChangeApplyStrategyOp string

const (
	noopStrategyOp    ClusterChangeApplyStrategyOp = ""
	UnknownStrategyOp ClusterChangeApplyStrategyOp = "unknown"
)

type ApplyStrategy interface {
	Op() ClusterChangeApplyStrategyOp
	Apply() error
}

type ClusterChangeOpts struct {
	ApplyIgnored bool
	Wait         bool
	WaitIgnored  bool

	AddOrUpdateChangeOpts
}

type ClusterChange struct {
	change              ctldiff.Change
	opts                ClusterChangeOpts
	identifiedResources ctlres.IdentifiedResources
	changeFactory       ctldiff.ChangeFactory
	changeSetFactory    ctldiff.ChangeSetFactory
	convergedResFactory ConvergedResourceFactory
	ui                  UI

	markedNeedsWaiting bool

	diffMaskRules []ctlconf.DiffMaskRule
}

var _ ChangeView = &ClusterChange{}

func NewClusterChange(change ctldiff.Change, opts ClusterChangeOpts,
	identifiedResources ctlres.IdentifiedResources,
	changeFactory ctldiff.ChangeFactory,
	changeSetFactory ctldiff.ChangeSetFactory,
	convergedResFactory ConvergedResourceFactory, ui UI,
	diffMaskRules []ctlconf.DiffMaskRule) *ClusterChange {

	return &ClusterChange{change, opts, identifiedResources,
		changeFactory, changeSetFactory, convergedResFactory, ui, false, diffMaskRules}
}

func (c *ClusterChange) ApplyOp() ClusterChangeApplyOp {
	if !c.opts.ApplyIgnored {
		if c.change.IsIgnored() {
			return ClusterChangeApplyOpNoop
		}
	}

	switch c.change.Op() {
	case ctldiff.ChangeOpAdd:
		return ClusterChangeApplyOpAdd
	case ctldiff.ChangeOpDelete:
		return ClusterChangeApplyOpDelete
	case ctldiff.ChangeOpUpdate:
		return ClusterChangeApplyOpUpdate
	case ctldiff.ChangeOpKeep:
		return ClusterChangeApplyOpNoop
	case ctldiff.ChangeOpExists:
		return ClusterChangeApplyOpExists
	case ctldiff.ChangeOpNoop:
		return ClusterChangeApplyOpNoop
	default:
		panic("Unknown change apply op")
	}
}

func (c *ClusterChange) WaitOp() ClusterChangeWaitOp {
	if !c.opts.Wait {
		return ClusterChangeWaitOpNoop
	}

	if !c.opts.WaitIgnored {
		if c.change.IsIgnored() {
			return ClusterChangeWaitOpNoop
		}
	}

	if _, found := c.Resource().Annotations()[disableWaitAnnKey]; found {
		return ClusterChangeWaitOpNoop
	}

	switch c.change.Op() {
	case ctldiff.ChangeOpAdd, ctldiff.ChangeOpUpdate:
		return ClusterChangeWaitOpOK

	case ctldiff.ChangeOpDelete:
		return ClusterChangeWaitOpDelete

	case ctldiff.ChangeOpKeep:
		// Return if this change was explicitly marked for waiting
		// as it may be a link in a dependency such as this:
		// A -> B -> C where,
		//   A is apply add
		//   B is apply noop and wait noop
		//   C is apply update and wait noop
		// Without marking B explicitly as needed to wait,
		// change in C may not be waited by A thru B.
		if c.markedNeedsWaiting {
			return ClusterChangeWaitOpOK
		}

		// TODO associated resources
		// If existing resource is not in a "done successful" state,
		// indicate that this will be something we need to wait for
		resState, _, err := c.convergedResFactory.New(c.change.ClusterOriginalResource(), nil).IsDoneApplying()
		if err != nil || !(resState.Done && resState.Successful) {
			return ClusterChangeWaitOpOK
		}
		return ClusterChangeWaitOpNoop

	case ctldiff.ChangeOpExists:
		return ClusterChangeWaitOpOK

	case ctldiff.ChangeOpNoop:
		return ClusterChangeWaitOpNoop

	default:
		panic("Unknown change wait op")
	}
}

func (c *ClusterChange) MarkNeedsWaiting() { c.markedNeedsWaiting = true }

func (c *ClusterChange) ApplyStrategyOp() (ClusterChangeApplyStrategyOp, error) {
	strategy, err := c.applyStrategy()
	if err != nil {
		return UnknownStrategyOp, err
	}
	return strategy.Op(), nil
}

func (c *ClusterChange) Apply() (bool, []string, error) {
	descMsgs := []string{c.ApplyDescription()}
	var retryable bool

	strategy, err := c.applyStrategy()
	if err != nil {
		return false, descMsgs, err
	}

	err = strategy.Apply()
	if err != nil {
		switch err.(type) {
		case ExistsChangeError:
			retryable = true
		default:
			retryable = ctlres.IsResourceChangeBlockedErr(err)
		}
	}

	if retryable {
		descMsgs = append(descMsgs, uiWaitMsgPrefix+"Retryable error: "+err.Error())
	}

	return retryable, descMsgs, c.applyErr(err)
}

func (c *ClusterChange) applyStrategy() (ApplyStrategy, error) {
	op := c.ApplyOp()

	switch op {
	case ClusterChangeApplyOpAdd, ClusterChangeApplyOpUpdate:
		return AddOrUpdateChange{
			c.change, c.identifiedResources, c.changeFactory,
			c.changeSetFactory, c.opts.AddOrUpdateChangeOpts, c.diffMaskRules}.ApplyStrategy()

	case ClusterChangeApplyOpDelete:
		return DeleteChange{c.change, c.identifiedResources}.ApplyStrategy()

	case ClusterChangeApplyOpNoop:
		return NoopStrategy{}, nil

	case ClusterChangeApplyOpExists:
		return ExistsChange{c.change, c.identifiedResources}.ApplyStrategy()

	default:
		return nil, fmt.Errorf("Unknown change apply operation: %s", op)
	}
}

func (c *ClusterChange) IsDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	state, descMsgs, err := c.isDoneApplying()
	primaryDescMsg := fmt.Sprintf("%s: %s", NewDoneApplyStateUI(state, err).State, c.WaitDescription())
	return state, append([]string{primaryDescMsg}, descMsgs...), err
}

func (c *ClusterChange) isDoneApplying() (ctlresm.DoneApplyState, []string, error) {
	op := c.WaitOp()

	switch op {
	case ClusterChangeWaitOpOK:
		return ReconcilingChange{c.change, c.identifiedResources, c.convergedResFactory}.IsDoneApplying()

	case ClusterChangeWaitOpDelete:
		return DeleteChange{c.change, c.identifiedResources}.IsDoneApplying()

	case ClusterChangeWaitOpNoop:
		return ctlresm.DoneApplyState{Done: true, Successful: true}, nil, nil

	default:
		return ctlresm.DoneApplyState{}, nil, fmt.Errorf("Unknown change wait operation: %s", op)
	}
}

func (c *ClusterChange) ApplyDescription() string {
	return fmt.Sprintf("%s %s", applyOpCodeUI[c.ApplyOp()], c.change.NewOrExistingResource().Description())
}

func (c *ClusterChange) WaitDescription() string {
	return fmt.Sprintf("%s %s", waitOpCodeUI[c.WaitOp()], c.change.NewOrExistingResource().Description())
}

func (c *ClusterChange) Resource() ctlres.Resource { return c.change.NewOrExistingResource() }

func (c *ClusterChange) ClusterOriginalResource() ctlres.Resource {
	return c.change.ClusterOriginalResource()
}

func (c *ClusterChange) ConfigurableTextDiff() *ctldiff.ConfigurableTextDiff {
	return c.change.ConfigurableTextDiff()
}

func (c *ClusterChange) applyErr(err error) error {
	if err == nil {
		return nil
	}

	hintMsg := ""
	hintableErrs := map[string]string{
		// TODO detect based on CRD content?
		"the server does not allow this method on the requested resource": "resource is possibly not namespaced but must be",
		"the server could not find the requested resource":                "resource is possibly namespaced but cannot be",
	}

	for errText, hintText := range hintableErrs {
		if strings.Contains(err.Error(), errText) {
			hintMsg = fmt.Sprintf(" (hint: %s)", hintText)
			break
		}
	}

	return fmt.Errorf("%s: %s%s", c.ApplyDescription(),
		uierrs.NewSemiStructuredError(err), hintMsg)
}

type NoopStrategy struct{}

func (s NoopStrategy) Op() ClusterChangeApplyStrategyOp { return noopStrategyOp }
func (s NoopStrategy) Apply() error                     { return nil }
