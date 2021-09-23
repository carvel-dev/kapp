// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/logger"
)

type ClusterChangeSetOpts struct {
	ApplyingChangesOpts
	WaitingChangesOpts
}

type ClusterChangeSet struct {
	changes              []ctldiff.Change
	opts                 ClusterChangeSetOpts
	clusterChangeFactory ClusterChangeFactory
	changeGroupBindings  []ctlconf.ChangeGroupBinding
	changeRuleBindings   []ctlconf.ChangeRuleBinding
	ui                   UI
	logger               logger.Logger
}

func NewClusterChangeSet(changes []ctldiff.Change, opts ClusterChangeSetOpts,
	clusterChangeFactory ClusterChangeFactory, changeGroupBindings []ctlconf.ChangeGroupBinding,
	changeRuleBindings []ctlconf.ChangeRuleBinding, ui UI, logger logger.Logger) ClusterChangeSet {

	return ClusterChangeSet{changes, opts, clusterChangeFactory,
		changeGroupBindings, changeRuleBindings, ui, logger.NewPrefixed("ClusterChangeSet")}
}

func (c ClusterChangeSet) Calculate() ([]*ClusterChange, *ctldgraph.ChangeGraph, error) {
	var wrappedClusterChanges []ctldgraph.ActualChange

	for _, change := range c.changes {
		clusterChange := c.clusterChangeFactory.NewClusterChange(change)
		wrappedClusterChanges = append(wrappedClusterChanges, wrappedClusterChange{clusterChange})
	}

	changesGraph, err := ctldgraph.NewChangeGraph(wrappedClusterChanges,
		c.changeGroupBindings, c.changeRuleBindings, c.logger)
	if err != nil {
		// Return graph for inspection
		return nil, changesGraph, err
	}

	changesGraph.AllMatching(func(change *ctldgraph.Change) bool {
		c.markChangesToWait(change)
		return false
	})

	// Prune out changes that are not involved with anything
	changesGraph.RemoveMatching(func(change *ctldgraph.Change) bool {
		clusterChange := change.Change.(wrappedClusterChange).ClusterChange

		return clusterChange.ApplyOp() == ClusterChangeApplyOpNoop &&
			clusterChange.WaitOp() == ClusterChangeWaitOpNoop
	})

	var clusterChanges []*ClusterChange

	for _, change := range changesGraph.All() {
		clusterChange := change.Change.(wrappedClusterChange).ClusterChange
		clusterChanges = append(clusterChanges, clusterChange)
	}

	return clusterChanges, changesGraph, nil
}

func (c ClusterChangeSet) markChangesToWait(change *ctldgraph.Change) bool {
	var needsWaiting bool
	for _, ch := range change.WaitingFor {
		if c.markChangesToWait(ch) {
			needsWaiting = true
			break
		}
	}
	if needsWaiting {
		change.Change.(wrappedClusterChange).MarkNeedsWaiting()
		return true
	}
	return change.Change.(wrappedClusterChange).WaitOp() != ClusterChangeWaitOpNoop
}

func (c ClusterChangeSet) Apply(changesGraph *ctldgraph.ChangeGraph) error {
	defer c.logger.DebugFunc("Apply").Finish()

	expectedNumChanges := len(changesGraph.All())

	blockedChanges := ctldgraph.NewBlockedChanges(changesGraph)
	applyingChanges := NewApplyingChanges(
		expectedNumChanges, c.opts.ApplyingChangesOpts, c.clusterChangeFactory, c.ui)
	waitingChanges := NewWaitingChanges(expectedNumChanges, c.opts.WaitingChangesOpts, c.ui)

	for {
		appliedChanges, err := applyingChanges.Apply(blockedChanges.Unblocked())
		if err != nil {
			return err
		}

		waitingChanges.Track(appliedChanges)

		if waitingChanges.IsEmpty() {
			err := applyingChanges.Complete()
			if err != nil {
				c.ui.Notify([]string{fmt.Sprintf("Blocked changes:\n%s\n", blockedChanges.WhyBlocked(blockedChanges.Blocked()))})
				return err
			}

			return waitingChanges.Complete()
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

func ClusterChangesAsChangeViews(changes []*ClusterChange) []ChangeView {
	var result []ChangeView
	for _, change := range changes {
		result = append(result, change)
	}
	return result
}

type wrappedClusterChange struct {
	*ClusterChange
}

func (c wrappedClusterChange) Op() ctldgraph.ActualChangeOp {
	op := c.ApplyOp()

	switch op {
	case ClusterChangeApplyOpAdd, ClusterChangeApplyOpUpdate:
		return ctldgraph.ActualChangeOpUpsert

	case ClusterChangeApplyOpDelete:
		return ctldgraph.ActualChangeOpDelete

	case ClusterChangeApplyOpNoop:
		return ctldgraph.ActualChangeOpNoop

	case ClusterChangeApplyOpExists:
		return ctldgraph.ActualChangeOpExists

	default:
		panic(fmt.Sprintf("Unknown change apply operation: %s", op))
	}
}

func (c wrappedClusterChange) WaitOp() ClusterChangeWaitOp {
	return c.ClusterChange.WaitOp()
}

func (c wrappedClusterChange) MarkNeedsWaiting() {
	c.ClusterChange.MarkNeedsWaiting()
}
