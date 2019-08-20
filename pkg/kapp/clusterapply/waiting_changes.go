package clusterapply

import (
	"fmt"
	"time"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
)

type WaitingChangesOpts struct {
	Timeout       time.Duration
	CheckInterval time.Duration
}

type WaitingChanges struct {
	numTotal       int // for ui
	numWaited      int // for ui
	trackedChanges []WaitingChange
	opts           WaitingChangesOpts
	ui             UI
}

type WaitingChange struct {
	Graph   *ctldgraph.Change
	Cluster *ClusterChange
}

func NewWaitingChanges(numTotal int, opts WaitingChangesOpts, ui UI) *WaitingChanges {
	return &WaitingChanges{numTotal, 0, nil, opts, ui}
}

func (c *WaitingChanges) Track(changes []WaitingChange) {
	c.trackedChanges = append(c.trackedChanges, changes...)
}

func (c *WaitingChanges) IsEmpty() bool {
	return len(c.trackedChanges) == 0
}

func (c *WaitingChanges) WaitForAny() ([]WaitingChange, error) {
	startTime := time.Now()

	for {
		c.ui.NotifySection("waiting on %d changes %s", len(c.trackedChanges), c.stats())

		var newInProgressChanges []WaitingChange
		var doneChanges []WaitingChange

		for _, change := range c.trackedChanges {
			desc := fmt.Sprintf("waiting on %s", change.Cluster.WaitDescription())

			state, descMsgs, err := change.Cluster.IsDoneApplying()
			c.ui.Notify(descMsgs)

			if err != nil {
				return nil, fmt.Errorf("%s: errored: %s", desc, err)
			}
			if state.Done {
				c.numWaited += 1
			}

			switch {
			case !state.Done:
				newInProgressChanges = append(newInProgressChanges, change)

			case state.Done && !state.Successful:
				msg := ""
				if len(state.Message) > 0 {
					msg += " (" + state.Message + ")"
				}
				return nil, fmt.Errorf("%s: finished unsuccessfully%s", desc, msg)

			case state.Done && state.Successful:
				doneChanges = append(doneChanges, change)
			}
		}

		c.trackedChanges = newInProgressChanges

		if len(c.trackedChanges) == 0 || len(doneChanges) > 0 {
			return doneChanges, nil
		}

		if time.Now().Sub(startTime) > c.opts.Timeout {
			return nil, fmt.Errorf("timed out waiting after %s", c.opts.Timeout)
		}

		time.Sleep(c.opts.CheckInterval)
	}
}

func (c *WaitingChanges) Complete() error {
	c.ui.NotifySection("waiting complete %s", c.stats())
	return nil
}

func (c *WaitingChanges) stats() string {
	return fmt.Sprintf("[%d/%d done]", c.numWaited, c.numTotal)
}
