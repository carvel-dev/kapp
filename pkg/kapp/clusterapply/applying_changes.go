package clusterapply

import (
	"fmt"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/util"
)

type ApplyingChangesOpts struct {
	Concurrency int
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

func (c *ApplyingChanges) Apply(allChanges []*ctldgraph.Change) ([]WaitingChange, error) {
	var nonAppliedChanges []*ctldgraph.Change

	for _, change := range allChanges {
		if !c.isApplied(change) {
			nonAppliedChanges = append(nonAppliedChanges, change)
		}
	}

	// Do not print applying message if no changes
	if len(nonAppliedChanges) == 0 {
		return nil, nil
	}

	c.ui.NotifySection("applying %d changes %s", len(nonAppliedChanges), c.stats())

	var result []WaitingChange
	applyErrsCh := make(chan error, len(nonAppliedChanges))

	// Throttle number of changes are applied concurrently
	// as it seems that client-go or api-server arent happy
	// with large number of updates going at once.
	// Example errors w/o throttling:
	// - "...: grpc: the client connection is closing (reason: )"
	// - "...: context canceled (reason: )"
	applyThrottle := util.NewThrottle(c.opts.Concurrency)

	for _, change := range nonAppliedChanges {
		c.markApplied(change)

		clusterChange := change.Change.(wrappedClusterChange).ClusterChange
		result = append(result, WaitingChange{change, clusterChange})

		go func() {
			applyThrottle.Take()
			defer applyThrottle.Done()

			// Print apply description as close to apply as possible to "show" apply progress
			c.ui.Notify([]string{clusterChange.ApplyDescription()})
			applyErrsCh <- clusterChange.Apply()
		}()
	}

	for i := 0; i < len(nonAppliedChanges); i++ {
		err := <-applyErrsCh
		if err != nil {
			return nil, err
		}
	}

	return result, nil
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
