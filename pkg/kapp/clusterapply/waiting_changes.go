// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"
	"time"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"github.com/k14s/kapp/pkg/kapp/util"
)

type WaitingChangesOpts struct {
	Timeout         time.Duration
	ResourceTimeout time.Duration
	CheckInterval   time.Duration
	Concurrency     int
}

type WaitingChanges struct {
	numTotal       int // for ui
	numWaited      int // for ui
	trackedChanges []WaitingChange
	opts           WaitingChangesOpts
	ui             UI
}

type WaitingChange struct {
	Graph     *ctldgraph.Change
	Cluster   *ClusterChange
	startTime time.Time
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

type waitResult struct {
	Change   WaitingChange
	State    ctlresm.DoneApplyState
	DescMsgs []string
	Err      error
}

func (c *WaitingChanges) WaitForAny() ([]WaitingChange, error) {
	startTime := time.Now()

	for {
		c.ui.NotifySection("waiting on %d changes %s", len(c.trackedChanges), c.stats())

		waitCh := make(chan waitResult, len(c.trackedChanges))
		waitThrottle := util.NewThrottle(c.opts.Concurrency)

		for _, change := range c.trackedChanges {
			change := change // copy

			go func() {
				waitThrottle.Take()
				defer waitThrottle.Done()

				state, descMsgs, err := change.Cluster.IsDoneApplying()
				// check for resource timeout
				if err == nil {
					if c.opts.ResourceTimeout != 0 && time.Now().Sub(change.startTime) > c.opts.ResourceTimeout {
						err = fmt.Errorf("Resource timed out waiting after %s", c.opts.ResourceTimeout)
					}
				}
				waitCh <- waitResult{Change: change, State: state, DescMsgs: descMsgs, Err: err}
			}()
		}

		var newInProgressChanges []WaitingChange
		var doneChanges []WaitingChange

		for i := 0; i < len(c.trackedChanges); i++ {
			result := <-waitCh
			change, state, descMsgs, err := result.Change, result.State, result.DescMsgs, result.Err

			desc := fmt.Sprintf("waiting on %s", change.Cluster.WaitDescription())
			c.ui.Notify(descMsgs)

			if err != nil {
				return nil, fmt.Errorf("%s: Errored: %s", desc, err)
			}
			if state.Done {
				c.numWaited++
			}

			switch {
			case !state.Done:
				newInProgressChanges = append(newInProgressChanges, change)

			case state.Done && !state.Successful:
				msg := ""
				if len(state.Message) > 0 {
					msg += " (" + state.Message + ")"
				}
				return nil, fmt.Errorf("%s: Finished unsuccessfully%s", desc, msg)

			case state.Done && state.Successful:
				doneChanges = append(doneChanges, change)
			}
		}

		c.trackedChanges = newInProgressChanges

		if len(c.trackedChanges) == 0 || len(doneChanges) > 0 {
			return doneChanges, nil
		}

		if time.Now().Sub(startTime) > c.opts.Timeout {
			return nil, fmt.Errorf("Timed out waiting after %s", c.opts.Timeout)
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
