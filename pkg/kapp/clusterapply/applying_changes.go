// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"
	"time"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/util"
)

type ApplyingChangesOpts struct {
	Timeout       time.Duration
	CheckInterval time.Duration
	Concurrency   int
}

type ApplyingChanges struct {
	numTotal             int // for ui
	opts                 ApplyingChangesOpts
	applied              map[*ctldgraph.Change]struct{}
	clusterChangeFactory ClusterChangeFactory
	ui                   UI
}

func NewApplyingChanges(numTotal int, opts ApplyingChangesOpts, clusterChangeFactory ClusterChangeFactory, ui UI) *ApplyingChanges {
	return &ApplyingChanges{numTotal, opts, map[*ctldgraph.Change]struct{}{}, clusterChangeFactory, ui}
}

type applyResult struct {
	Change        *ctldgraph.Change
	ClusterChange *ClusterChange
	DescMsgs      []string
	Retryable     bool
	Err           error
}

func (c *ApplyingChanges) Apply(allChanges []*ctldgraph.Change) ([]WaitingChange, error) {
	startTime := time.Now()

	for {
		nonAppliedChanges := c.nonAppliedChanges(allChanges)
		if len(nonAppliedChanges) == 0 {
			// Do not print applying message if no changes
			return nil, nil
		}

		c.ui.NotifySection("applying %d changes %s", len(nonAppliedChanges), c.stats())

		// Throttle number of changes are applied concurrently
		// as it seems that client-go or api-server arent happy
		// with large number of updates going at once.
		// Example errors w/o throttling:
		// - "...: grpc: the client connection is closing (reason: )"
		// - "...: context canceled (reason: )"
		applyThrottle := util.NewThrottle(c.opts.Concurrency)
		applyCh := make(chan applyResult, len(nonAppliedChanges))

		for _, change := range nonAppliedChanges {
			change := change // copy

			go func() {
				applyThrottle.Take()
				defer applyThrottle.Done()

				clusterChange := change.Change.(wrappedClusterChange).ClusterChange
				retryable, descMsgs, err := clusterChange.Apply()

				applyCh <- applyResult{
					Change:        change,
					ClusterChange: clusterChange,
					DescMsgs:      descMsgs,
					Retryable:     retryable,
					Err:           err,
				}
			}()
		}

		var appliedChanges []WaitingChange
		var lastErr error

		for i := 0; i < len(nonAppliedChanges); i++ {
			result := <-applyCh

			c.ui.Notify(result.DescMsgs)

			if result.Err != nil {
				lastErr = result.Err
				if result.Retryable {
					continue
				}
				return nil, result.Err
			}

			c.markApplied(result.Change)
			appliedChanges = append(appliedChanges, WaitingChange{result.Change, result.ClusterChange, time.Time{}})
		}

		if len(appliedChanges) > 0 {
			return appliedChanges, nil
		}

		if time.Now().Sub(startTime) > c.opts.Timeout {
			return nil, fmt.Errorf("Timed out waiting after %s: Last error: %s", c.opts.Timeout, lastErr)
		}

		time.Sleep(c.opts.CheckInterval)
	}
}

func (c *ApplyingChanges) Complete() error {
	// Sanity check that we applied all changes
	if c.numTotal != c.numApplied() {
		return fmt.Errorf("Internal inconsistency: did not apply all changes: %d != %d",
			c.numTotal, c.numApplied())
	}

	c.ui.NotifySection("applying complete %s", c.stats())
	return nil
}

func (c *ApplyingChanges) nonAppliedChanges(allChanges []*ctldgraph.Change) []*ctldgraph.Change {
	var result []*ctldgraph.Change
	for _, change := range allChanges {
		if !c.isApplied(change) {
			result = append(result, change)
		}
	}
	return result
}

func (c *ApplyingChanges) isApplied(change *ctldgraph.Change) bool {
	_, found := c.applied[change]
	return found
}

func (c *ApplyingChanges) markApplied(change *ctldgraph.Change) {
	c.applied[change] = struct{}{}
}

func (c *ApplyingChanges) numApplied() int { return len(c.applied) }

func (c *ApplyingChanges) stats() string {
	return fmt.Sprintf("[%d/%d done]", c.numApplied(), c.numTotal)
}
