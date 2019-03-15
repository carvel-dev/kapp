package clusterapply

import (
	"fmt"
	"strings"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

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
	ui                  UI
}

func NewClusterChange(change ctldiff.Change, opts ClusterChangeOpts,
	identifiedResources ctlres.IdentifiedResources, changeFactory ctldiff.ChangeFactory, ui UI) ClusterChange {

	return ClusterChange{change, opts, identifiedResources, changeFactory, ui}
}

func (c ClusterChange) ApplyOp() ApplyOp {
	if !c.opts.ApplyIgnored {
		if c.change.IsIgnored() {
			return noopApplyOp
		}
	}

	op := c.change.Op()

	switch op {
	case ctldiff.ChangeOpAdd, ctldiff.ChangeOpUpdate:
		return ApplyOp{func() error {
			return c.applyErr(AddOrUpdateChange{
				c.change, c.identifiedResources, c.changeFactory,
				c.opts.AddOrUpdateChangeOpts, c.ui}.Apply())
		}}

	case ctldiff.ChangeOpDelete:
		return ApplyOp{func() error {
			return c.applyErr(DeleteChange{c.change, c.identifiedResources}.Apply())
		}}

	case ctldiff.ChangeOpKeep:
		return noopApplyOp

	default:
		return ApplyOp{func() error {
			return fmt.Errorf("Unknown change operation: %s", op)
		}}
	}
}

func (c ClusterChange) IsDoneApplyingOp() DoneWaitingOp {
	if !c.opts.Wait {
		return noopDoneWaitingOp
	}

	if !c.opts.WaitIgnored {
		if c.change.IsIgnored() {
			return noopDoneWaitingOp
		}
	}

	op := c.change.Op()

	// TODO CRD status conditions
	// TODO jobs, pod status?

	switch op {
	case ctldiff.ChangeOpAdd, ctldiff.ChangeOpUpdate:
		return DoneWaitingOp{AddOrUpdateChange{
			c.change, c.identifiedResources, c.changeFactory,
			c.opts.AddOrUpdateChangeOpts, c.ui}.IsDoneApplying}

	case ctldiff.ChangeOpDelete:
		return DoneWaitingOp{DeleteChange{c.change, c.identifiedResources}.IsDoneApplying}

	case ctldiff.ChangeOpKeep:
		return noopDoneWaitingOp

	default:
		return DoneWaitingOp{func() (ctlresm.DoneApplyState, error) {
			return ctlresm.DoneApplyState{}, fmt.Errorf("Unknown change operation: %s", op)
		}}
	}
}

func (c ClusterChange) ApplyDescription() string {
	if c.ApplyOp().IsNoop() {
		return ""
	}
	return fmt.Sprintf("%s %s", c.change.Op(), c.change.NewOrExistingResource().Description())
}

func (c ClusterChange) WaitDescription() string {
	if c.IsDoneApplyingOp().IsNoop() {
		return ""
	}
	return fmt.Sprintf("%s %s", c.change.Op(), c.change.NewOrExistingResource().Description())
}

func (c ClusterChange) applyErr(err error) error {
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

	return fmt.Errorf("Applying change operation '%s' to '%s': %s%s",
		c.change.Op(), c.change.NewOrExistingResource().Description(), err, hintMsg)
}

var (
	noopApplyOp       = ApplyOp{}
	noopDoneWaitingOp = DoneWaitingOp{}
)

type ApplyOp struct {
	f func() error
}

func (op ApplyOp) IsNoop() bool { return op.f == nil }

func (op ApplyOp) Execute() error {
	if op.IsNoop() {
		return nil
	}
	return op.f()
}

type DoneWaitingOp struct {
	f func() (ctlresm.DoneApplyState, error)
}

func (op DoneWaitingOp) IsNoop() bool { return op.f == nil }

func (op DoneWaitingOp) Execute() (ctlresm.DoneApplyState, error) {
	if op.IsNoop() {
		return ctlresm.DoneApplyState{Done: true, Successful: true}, nil
	}
	return op.f()
}
