package clusterapply

import (
	"fmt"
	"strings"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
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
)

type ClusterChangeWaitOp string

const (
	ClusterChangeWaitOpOK     ClusterChangeWaitOp = "ok"
	ClusterChangeWaitOpDelete ClusterChangeWaitOp = "delete"
	ClusterChangeWaitOpNoop   ClusterChangeWaitOp = "noop"
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

	markedNeedsWaiting bool
}

var _ ChangeView = &ClusterChange{}

func NewClusterChange(change ctldiff.Change, opts ClusterChangeOpts,
	identifiedResources ctlres.IdentifiedResources, changeFactory ctldiff.ChangeFactory, ui UI) *ClusterChange {

	return &ClusterChange{change, opts, identifiedResources, changeFactory, ui, false}
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
		existingResState, existingErr := NewConvergedResource(c.change.ExistingResource(), nil).IsDoneApplying(&noopUI{})
		if existingErr != nil || !(existingResState.Done && existingResState.Successful) {
			return ClusterChangeWaitOpOK
		}
		return ClusterChangeWaitOpNoop

	default:
		panic("Unknown change wait op")
	}
}

func (c *ClusterChange) MarkNeedsWaiting() { c.markedNeedsWaiting = true }

func (c *ClusterChange) Apply() error {
	op := c.ApplyOp()

	switch op {
	case ClusterChangeApplyOpAdd, ClusterChangeApplyOpUpdate:
		return c.applyErr(AddOrUpdateChange{
			c.change, c.identifiedResources, c.changeFactory,
			c.opts.AddOrUpdateChangeOpts, c.ui}.Apply())

	case ClusterChangeApplyOpDelete:
		return c.applyErr(DeleteChange{c.change, c.identifiedResources}.Apply())

	case ClusterChangeApplyOpNoop:
		return nil

	default:
		return fmt.Errorf("Unknown change apply operation: %s", op)
	}
}

func (c *ClusterChange) IsDoneApplying() (ctlresm.DoneApplyState, error) {
	op := c.WaitOp()

	switch op {
	case ClusterChangeWaitOpOK:
		return AddOrUpdateChange{
			c.change, c.identifiedResources, c.changeFactory,
			c.opts.AddOrUpdateChangeOpts, c.ui}.IsDoneApplying()

	case ClusterChangeWaitOpDelete:
		return DeleteChange{c.change, c.identifiedResources}.IsDoneApplying()

	case ClusterChangeWaitOpNoop:
		return ctlresm.DoneApplyState{Done: true, Successful: true}, nil

	default:
		return ctlresm.DoneApplyState{}, fmt.Errorf("Unknown change wait operation: %s", op)
	}
}

func (c *ClusterChange) ApplyDescription() string {
	op := c.ApplyOp()
	switch op {
	case ClusterChangeApplyOpNoop:
		return ""
	default:
		return fmt.Sprintf("%s %s", applyOpCodeUI[op], c.change.NewOrExistingResource().Description())
	}
}

func (c *ClusterChange) WaitDescription() string {
	op := c.WaitOp()
	switch op {
	case ClusterChangeWaitOpNoop:
		return ""
	default:
		return fmt.Sprintf("%s %s", waitOpCodeUI[op], c.change.NewOrExistingResource().Description())
	}
}

func (c *ClusterChange) Resource() ctlres.Resource  { return c.change.NewOrExistingResource() }
func (c *ClusterChange) TextDiff() ctldiff.TextDiff { return c.change.TextDiff() }
func (c *ClusterChange) IsIgnored() bool            { return c.change.IsIgnored() }
func (c *ClusterChange) IgnoredReason() string      { return c.change.IgnoredReason() }

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

	return fmt.Errorf("Applying change operation '%s' to '%s': %s%s",
		c.change.Op(), c.change.NewOrExistingResource().Description(), err, hintMsg)
}

type noopUI struct{}

func (b *noopUI) NotifySection(msg string, args ...interface{}) {}
func (b *noopUI) Notify(msg string, args ...interface{})        {}
