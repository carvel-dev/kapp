package clusterapply

import (
	"fmt"
	"time"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
)

type WaitingChanges struct {
	numTotal       int // for ui
	numWaited      int // for ui
	trackedChanges []WaitingChange
	opts           ClusterChangeSetOpts
	ui             UI
}

type WaitingChange struct {
	Graph   *ctldgraph.Change
	Cluster ClusterChange
}

func NewWaitingChanges(numTotal int, opts ClusterChangeSetOpts, ui UI) *WaitingChanges {
	return &WaitingChanges{numTotal, 0, nil, opts, ui}
}

func (c *WaitingChanges) Track(changes []WaitingChange) {
	c.numWaited += len(changes)
	c.trackedChanges = append(c.trackedChanges, changes...)
}

func (c *WaitingChanges) IsEmpty() bool {
	return len(c.trackedChanges) == 0
}

func (c *WaitingChanges) WaitForAny() ([]WaitingChange, error) {
	startTime := time.Now()

	for {
		c.ui.Notify("")
		c.ui.Notify(uiSepPrefix+"waiting on %d changes [%d/%d]",
			len(c.trackedChanges), c.numWaited, c.numTotal)

		var newInProgressChanges []WaitingChange
		var doneChanges []WaitingChange

		for _, change := range c.trackedChanges {
			desc := change.Cluster.WaitDescription()
			if len(desc) > 0 {
				c.ui.Notify("waiting on %s", desc)
			}

			state, err := change.Cluster.IsDoneApplyingOp().Execute()
			if err != nil {
				return nil, fmt.Errorf("waiting on %s errored: %s", desc, err)
			}

			switch {
			case !state.Done:
				newInProgressChanges = append(newInProgressChanges, change)

			case state.Done && !state.Successful:
				msg := ""
				if len(state.Message) > 0 {
					msg += " (" + state.Message + ")"
				}
				return nil, fmt.Errorf("waiting on %s: finished unsuccessfully%s", desc, msg)

			case state.Done && state.Successful:
				doneChanges = append(doneChanges, change)
			}
		}

		c.trackedChanges = newInProgressChanges

		if len(c.trackedChanges) == 0 || len(doneChanges) > 0 {
			return doneChanges, nil
		}

		if time.Now().Sub(startTime) > c.opts.WaitTimeout {
			return nil, fmt.Errorf("timed out waiting after %s", c.opts.WaitTimeout)
		}

		time.Sleep(c.opts.WaitCheckInterval)
	}
}
