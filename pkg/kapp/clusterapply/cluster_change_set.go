package clusterapply

import (
	"fmt"
	"time"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
)

type ClusterChangeSetOpts struct {
	WaitTimeout       time.Duration
	WaitCheckInterval time.Duration
}

type ClusterChangeSet struct {
	changes              []ctldiff.Change
	opts                 ClusterChangeSetOpts
	clusterChangeFactory ClusterChangeFactory
	ui                   UI
}

func NewClusterChangeSet(changes []ctldiff.Change, opts ClusterChangeSetOpts,
	clusterChangeFactory ClusterChangeFactory, ui UI) ClusterChangeSet {

	return ClusterChangeSet{changes, opts, clusterChangeFactory, ui}
}

func (c ClusterChangeSet) Calculate() []ChangeView {
	var result []ChangeView
	for _, change := range c.changes {
		clusterChange := c.clusterChangeFactory.NewClusterChange(change)
		result = append(result, clusterChange)
	}
	return result
}

func (c ClusterChangeSet) Apply() error {
	changesGraph, err := ctldgraph.NewChangeGraph(c.changes)
	if err != nil {
		return err
	}

	blockedChanges := ctldgraph.NewBlockedChanges(changesGraph)
	applyingChanges := NewApplyingChanges(len(c.changes), c.clusterChangeFactory, c.ui)
	waitingChanges := NewWaitingChanges(len(c.changes), c.opts, c.ui)

	for {
		appliedChanges, err := applyingChanges.Apply(blockedChanges.Unblocked())
		if err != nil {
			return err
		}

		waitingChanges.Track(appliedChanges)

		if waitingChanges.IsEmpty() {
			// Sanity check that we applied all changes
			expectedChangesLen := len(c.changes)
			appliedChangesLen := applyingChanges.NumApplied()

			if expectedChangesLen != appliedChangesLen {
				fmt.Printf("%s\n", blockedChanges.WhyBlocked(blockedChanges.Blocked()))
				return fmt.Errorf("Internal inconsistency: did not apply all changes: %d != %d",
					expectedChangesLen, appliedChangesLen)
			}

			c.ui.NotifySection("changes applied")
			return nil
		}

		doneChanges, err := waitingChanges.WaitForAny()
		if err != nil {
			return err
		}

		for _, change := range doneChanges {
			blockedChanges.Unblock(change.Graph)
		}
	}
}
