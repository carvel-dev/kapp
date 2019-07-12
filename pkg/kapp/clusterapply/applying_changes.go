package clusterapply

import (
	"sync"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
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

	c.ui.NotifySection("applying %d changes [%d/%d]",
		len(nonAppliedChanges), c.NumApplied()+len(nonAppliedChanges), c.numTotal)

	var wg sync.WaitGroup
	var result []WaitingChange
	applyErrCh := make(chan error, len(nonAppliedChanges))

	for _, change := range nonAppliedChanges {
		c.markApplied(change)
		clusterChange := c.clusterChangeFactory.NewClusterChange(change.Change)

		desc := clusterChange.ApplyDescription()
		if len(desc) > 0 {
			c.ui.Notify("%s", desc)
		}

		wg.Add(1)

		go func() {
			defer func() { wg.Done() }()

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

func (c *ApplyingChanges) NumApplied() int {
	return len(c.applied)
}

func (c *ApplyingChanges) isApplied(change *ctldgraph.Change) bool {
	_, found := c.applied[change]
	return found
}

func (c *ApplyingChanges) markApplied(change *ctldgraph.Change) {
	c.applied[change] = struct{}{}
}
