// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
	"sort"
	"strings"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	"github.com/k14s/kapp/pkg/kapp/logger"
)

type ChangeGraph struct {
	changes []*Change
	logger  logger.Logger
}

func NewChangeGraph(changes []ActualChange,
	changeGroupBindings []ctlconf.ChangeGroupBinding,
	changeRuleBindings []ctlconf.ChangeRuleBinding,
	logger logger.Logger) (*ChangeGraph, error) {

	logger = logger.NewPrefixed("ChangeGraph")
	defer logger.DebugFunc("NewChangeGraph").Finish()

	graphChanges := []*Change{}

	for _, change := range changes {
		graphChanges = append(graphChanges, &Change{
			Change:              change,
			changeGroupBindings: changeGroupBindings,
			changeRuleBindings:  changeRuleBindings,
		})
	}
	graph := &ChangeGraph{graphChanges, logger}

	err := graph.buildEdges(
		func(rule ChangeRule) bool { return !rule.IgnoreIfCyclical },
		func(_, _ *Change) bool { return true },
	)
	if err != nil {
		return graph, fmt.Errorf("Change graph: Calculating required deps: %s", err)
	}

	err = graph.checkCycles()
	if err != nil {
		// Return graph for inspection
		return graph, err
	}

	// At this point it's guranteed that there are no cycles
	// Start adding optional rules but only if they do not introduce
	// new cycles. For example, given
	//   A -> B -> C
	// if we try to add
	//   C -> A
	// cycle will be formed. To check that quickly it's only necessary
	// to check if one can get to C from A, hence, C -> A is rejected.

	err = graph.buildEdges(
		func(rule ChangeRule) bool { return rule.IgnoreIfCyclical },
		func(graphChange, matchedChange *Change) bool {
			return graphChange != matchedChange &&
				!graphChange.IsDirectlyWaitingFor(matchedChange) &&
				!matchedChange.IsTransitivelyWaitingFor(graphChange)
		},
	)
	if err != nil {
		return graph, fmt.Errorf("Change graph: Calculating optional deps: %s", err)
	}

	graph.dedup()

	// Double check cycles again
	return graph, graph.checkCycles()
}

type sortedRule struct {
	*Change
	ChangeRule
}

func (g *ChangeGraph) buildEdges(allowRule func(ChangeRule) bool,
	allowChange func(*Change, *Change) bool) error {

	defer g.logger.DebugFunc("buildEdges").Finish()

	var sortedRules []sortedRule

	for _, graphChange := range g.changes {
		rules, err := graphChange.ApplicableRules()
		if err != nil {
			return err
		}

		for _, rule := range rules {
			if allowRule(rule) {
				sortedRules = append(sortedRules, sortedRule{graphChange, rule})
			}
		}
	}

	// Since some rules may conflict with other rules (cause cycles)
	// we need to order rules so that they are added deterministically
	sort.SliceStable(sortedRules, func(i, j int) bool {
		// Higher weighted rules come first
		return sortedRules[i].ChangeRule.weight > sortedRules[j].ChangeRule.weight
	})

	for _, sr := range sortedRules {
		matchedChanges, err := Changes(g.changes).MatchesRule(sr.ChangeRule, sr.Change)
		if err != nil {
			return err
		}

		switch {
		case sr.ChangeRule.Order == ChangeRuleOrderAfter:
			for _, matchedChange := range matchedChanges {
				if allowChange(sr.Change, matchedChange) {
					sr.Change.WaitingFor = append(sr.Change.WaitingFor, matchedChange)
				}
			}

		case sr.ChangeRule.Order == ChangeRuleOrderBefore:
			for _, matchedChange := range matchedChanges {
				if allowChange(matchedChange, sr.Change) {
					matchedChange.WaitingFor = append(matchedChange.WaitingFor, sr.Change)
				}
			}

		default:
			panic("Unknown change rule order")
		}
	}

	return nil
}

func (g *ChangeGraph) All() []*Change {
	return g.AllMatching(func(_ *Change) bool { return true })
}

func (g *ChangeGraph) Linearized() ([][]*Change, []*Change) {
	var resultLinearized [][]*Change
	var resultBlocked []*Change

	recordedChanges := map[*Change]struct{}{}
	blockedChanges := NewBlockedChanges(g)
	lastBlockedChanges := 0

	for {
		unblocked := blockedChanges.Unblocked()
		blocked := blockedChanges.Blocked()

		var sectionLinearized []*Change
		for _, unblockedChange := range unblocked {
			if _, found := recordedChanges[unblockedChange]; !found {
				recordedChanges[unblockedChange] = struct{}{}
				blockedChanges.Unblock(unblockedChange)
				sectionLinearized = append(sectionLinearized, unblockedChange)
			}
		}
		resultLinearized = append(resultLinearized, sectionLinearized)

		if len(blocked) == 0 || len(blocked) == lastBlockedChanges {
			for _, blockedChange := range blocked {
				resultBlocked = append(resultBlocked, blockedChange)
			}
			return resultLinearized, resultBlocked
		}

		lastBlockedChanges = len(blocked)
	}
}

func (g *ChangeGraph) AllMatching(matchFunc func(*Change) bool) []*Change {
	var result []*Change
	// Need to do this _only_ at the first level since
	// all changes are included at the top level
	for _, change := range g.changes {
		if matchFunc(change) {
			result = append(result, change)
		}
	}
	return result
}

func (g *ChangeGraph) RemoveMatching(matchFunc func(*Change) bool) {
	var result []*Change
	// Need to do this _only_ at the first level since
	// all changes are included at the top level
	for _, change := range g.changes {
		if !matchFunc(change) {
			result = append(result, change)
		}
	}
	g.changes = result
}

func (g *ChangeGraph) Print() {
	fmt.Printf("%s", g.PrintStr())
}

func (g *ChangeGraph) PrintStr() string {
	return g.printChanges(g.changes, map[*Change]bool{}, "")
}

func (g *ChangeGraph) PrintLinearizedStr() string {
	linearizedChangeSections, blockedChanges := g.Linearized()

	var result []string

	for _, changes := range linearizedChangeSections {
		var section []string
		for _, change := range changes {
			section = append(section, change.Description())
		}
		result = append(result, strings.Join(section, "\n"))
	}

	if len(blockedChanges) > 0 {
		var section []string
		for _, change := range blockedChanges {
			section = append(section, change.Description())
		}
		result = append(result, "...more blocked...\n"+strings.Join(section, "\n"))
	}

	return strings.Join(result, "\n---\n")
}

func (g *ChangeGraph) printChanges(changes []*Change,
	visitedChanges map[*Change]bool, indent string) string {

	var result string

	for _, change := range changes {
		result += fmt.Sprintf("%s%s\n", indent, change.Description())

		if _, found := visitedChanges[change]; !found {
			visitedChanges[change] = true
			result += g.printChanges(change.WaitingFor, visitedChanges, indent+"  ")
			delete(visitedChanges, change)
		} else {
			result += indent + "cycle found\n"
		}
	}

	return result
}

func (g *ChangeGraph) dedup() {
	for _, rootChange := range g.changes {
		seenWaitingFor := map[*Change]struct{}{}
		newWaitingFor := []*Change{}

		for _, change := range rootChange.WaitingFor {
			if _, ok := seenWaitingFor[change]; !ok {
				seenWaitingFor[change] = struct{}{}
				newWaitingFor = append(newWaitingFor, change)
			}
		}

		rootChange.WaitingFor = newWaitingFor
	}
}

// Implements depth-first search:
// https://en.wikipedia.org/wiki/Topological_sorting#Depth-first_search
func (g *ChangeGraph) checkCycles() error {
	defer g.logger.DebugFunc("checkCycles").Finish()

	markedTemp := map[*Change]struct{}{}
	markedPerm := map[*Change]struct{}{}
	unmarked := []*Change{}

	for _, change := range g.changes {
		unmarked = append(unmarked, change)
	}

	for len(unmarked) > 0 {
		nodeN := unmarked[0]
		unmarked = unmarked[1:]
		err := g.checkCyclesVisit(nodeN, markedTemp, markedPerm)
		if err != nil {
			return fmt.Errorf("Detected cycle while ordering changes: [%s] %s",
				nodeN.Change.Resource().Description(), err)
		}
	}

	return nil
}

func (g *ChangeGraph) checkCyclesVisit(nodeN *Change, markedTemp, markedPerm map[*Change]struct{}) error {
	if _, found := markedPerm[nodeN]; found {
		return nil
	}
	if _, found := markedTemp[nodeN]; found {
		return fmt.Errorf("(found repeated: %s)", nodeN.Change.Resource().Description())
	}
	markedTemp[nodeN] = struct{}{}

	for _, nodeM := range nodeN.WaitingFor {
		err := g.checkCyclesVisit(nodeM, markedTemp, markedPerm)
		if err != nil {
			return fmt.Errorf("-> [%s] %s", nodeM.Change.Resource().Description(), err)
		}
	}

	delete(markedTemp, nodeN)
	markedPerm[nodeN] = struct{}{}
	return nil
}
