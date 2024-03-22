// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
)

type BlockedChanges struct {
	graph     *ChangeGraph
	unblocked map[*Change]struct{}
}

func NewBlockedChanges(graph *ChangeGraph) *BlockedChanges {
	return &BlockedChanges{graph, map[*Change]struct{}{}}
}

func (c *BlockedChanges) Unblocked() []*Change {
	return c.graph.AllMatching(c.isUnblocked)
}

func (c *BlockedChanges) Blocked() []*Change {
	return c.graph.AllMatching(func(change *Change) bool { return !c.isUnblocked(change) })
}

func (c *BlockedChanges) WhyBlocked(changes []*Change) string {
	var result string
	for _, change := range changes {
		result += fmt.Sprintf("%s\n", change.Change.Resource().Description())
		for _, childChange := range change.WaitingFor {
			if c.isBlocked(childChange) {
				result += fmt.Sprintf("  [blocked] %s\n", childChange.Change.Resource().Description())
			}
		}
	}
	return result
}

func (c *BlockedChanges) Unblock(change *Change) {
	c.unblocked[change] = struct{}{}
}

func (c *BlockedChanges) isUnblocked(change *Change) bool {
	for _, childChange := range change.WaitingFor {
		if c.isBlocked(childChange) {
			return false
		}
	}
	return true
}

func (c *BlockedChanges) isBlocked(change *Change) bool {
	_, found := c.unblocked[change]
	return !found
}
