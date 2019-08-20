package clusterapply

import (
	"fmt"
	"sync"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/util"
)

type ApplyingChanges struct {
	numTotal             int // for ui
	applied              map[*ctldgraph.Change]struct{}
	clusterChangeFactory ClusterChangeFactory
	ui                   UI
}

func NewApplyingChanges(numTotal int, clusterChangeFactory ClusterChangeFactory, ui UI) *ApplyingChanges {
	return &ApplyingChanges{numTotal, map[*ctldgraph.Change]struct{}{}, clusterChangeFactory, ui}
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

	var wg sync.WaitGroup
	var result []WaitingChange
	applyErrCh := make(chan error, len(nonAppliedChanges))

	// Throttle number of changes are applied concurrently
	// as it seems that client-go or api-server arent happy
	// with large number of updates going at once.
	// Example errors w/o throttling:
	// - "...: grpc: the client connection is closing (reason: )"
	// - "...: context canceled (reason: )"
	applyThrottle := util.NewThrottle(5)

	for _, change := range nonAppliedChanges {
		c.markApplied(change)
		clusterChange := change.Change.(wrappedClusterChange).ClusterChange

		c.ui.Notify([]string{clusterChange.ApplyDescription()})
		wg.Add(1)

		go func() {
			defer func() { wg.Done() }()

			applyThrottle.Take()
			defer applyThrottle.Done()

			err := clusterChange.Apply()
			applyErrCh <- err
		}()

		result = append(result, WaitingChange{change, clusterChange})
	}

	wg.Wait()
	close(applyErrCh)

	for err := range applyErrCh {
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
