// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"
	"strings"
	"time"

	ctldgraph "carvel.dev/kapp/pkg/kapp/diffgraph"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
	"carvel.dev/kapp/pkg/kapp/util"
	uierrs "github.com/cppforlife/go-cli-ui/errors"
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
	exitOnError    bool
}

type WaitingChange struct {
	Graph     *ctldgraph.Change
	Cluster   *ClusterChange
	startTime time.Time
}

func NewWaitingChanges(numTotal int, opts WaitingChangesOpts, ui UI, exitOnError bool) *WaitingChanges {
	return &WaitingChanges{numTotal, 0, nil, opts, ui, exitOnError}
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

func (c *WaitingChanges) WaitForAny() ([]WaitingChange, []string, error) {
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
		var unsuccessfulChangeDesc []string

		for i := 0; i < len(c.trackedChanges); i++ {
			result := <-waitCh
			change, state, descMsgs, err := result.Change, result.State, result.DescMsgs, result.Err

			desc := fmt.Sprintf("waiting on %s", change.Cluster.WaitDescription())
			c.ui.Notify(descMsgs)

			if err != nil {
				err = fmt.Errorf("%s: Errored: %w", desc, err)
				if c.exitOnError {
					return nil, nil, err
				}
				unsuccessfulChangeDesc = append(unsuccessfulChangeDesc, err.Error())
				continue
			}
			if state.Done {
				c.numWaited++
			}

			switch {
			case !state.Done:
				newInProgressChanges = append(newInProgressChanges, change)

				if state.UnblockChanges {
					doneChanges = append(doneChanges, change)
				}

			case state.Done && !state.Successful:
				msg := ""
				if len(state.Message) > 0 {
					msg = ": " + state.Message
				}
				err := fmt.Errorf("%s: Finished waiting unsuccessfully%s", desc, msg)
				if c.exitOnError {
					return nil, nil, err
				}
				unsuccessfulChangeDesc = append(unsuccessfulChangeDesc, err.Error())

			case state.Done && state.Successful:
				doneChanges = append(doneChanges, change)
			}
		}

		c.trackedChanges = newInProgressChanges

		if len(c.trackedChanges) == 0 || len(doneChanges) > 0 || len(unsuccessfulChangeDesc) > 0 {
			return doneChanges, unsuccessfulChangeDesc, nil
		}

		if time.Now().Sub(startTime) > c.opts.Timeout {
			var trackedResourcesDesc []string
			for _, change := range c.trackedChanges {
				trackedResourcesDesc = append(trackedResourcesDesc, change.Cluster.Resource().Description())
			}
			return nil, unsuccessfulChangeDesc, uierrs.NewSemiStructuredError(fmt.Errorf("Timed out waiting after %s for resources: [%s]", c.opts.Timeout, strings.Join(trackedResourcesDesc, ", ")))
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
