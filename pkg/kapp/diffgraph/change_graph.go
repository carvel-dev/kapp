package diffgraph

import (
	"fmt"
	"strings"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	"github.com/k14s/kapp/pkg/kapp/logger"
)

type ChangeGraph struct {
	changes []*Change
	logger  logger.Logger
}

func NewChangeGraph(changes []ActualChange,
	additionalChangeGroups []ctlconf.AdditionalChangeGroup,
	additionalChangeRules []ctlconf.AdditionalChangeRule,
	logger logger.Logger) (*ChangeGraph, error) {

	logger = logger.NewPrefixed("ChangeGraph")
	defer logger.DebugFunc("NewChangeGraph").Finish()

	graphChanges := []*Change{}

	for _, change := range changes {
		graphChanges = append(graphChanges, &Change{
			Change:                 change,
			additionalChangeGroups: additionalChangeGroups,
			additionalChangeRules:  additionalChangeRules,
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

func (g *ChangeGraph) buildEdges(allowRule func(ChangeRule) bool,
	allowChange func(*Change, *Change) bool) error {

	defer g.logger.DebugFunc("buildEdges").Finish()

	for _, graphChange := range g.changes {
		rules, err := graphChange.ApplicableRules()
		if err != nil {
			return err
		}

		for _, rule := range rules {
			if !allowRule(rule) {
				continue
			}

			matchedChanges, err := Changes(g.changes).MatchesRule(rule, graphChange)
			if err != nil {
				return err
			}

			switch {
			case rule.Order == ChangeRuleOrderAfter:
				for _, matchedChange := range matchedChanges {
					if allowChange(graphChange, matchedChange) {
						graphChange.WaitingFor = append(graphChange.WaitingFor, matchedChange)
					}
				}

			case rule.Order == ChangeRuleOrderBefore:
				for _, matchedChange := range matchedChanges {
					if allowChange(matchedChange, graphChange) {
						matchedChange.WaitingFor = append(matchedChange.WaitingFor, graphChange)
					}
				}

			default:
				panic("Unknown change rule order")
			}
		}
	}
	return nil
}

func (g *ChangeGraph) All() []*Change {
	return g.AllMatching(func(_ *Change) bool { return true })
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

func (g *ChangeGraph) PrintLinearizedStr() string {
	result := []string{}
	recordedChanges := map[*Change]struct{}{}
	blockedChanges := NewBlockedChanges(g)
	lastBlockedChanges := 0

	for {
		unblocked := blockedChanges.Unblocked()
		blocked := blockedChanges.Blocked()

		if len(blocked) == lastBlockedChanges {
			var section []string
			for _, blockedChange := range blocked {
				section = append(section, blockedChange.Description())
			}
			result = append(result, "...more blocked...\n"+strings.Join(section, "\n"))
			return strings.Join(result, "\n---\n")
		}
		lastBlockedChanges = len(blocked)

		var section []string
		for _, unblockedChange := range unblocked {
			if _, found := recordedChanges[unblockedChange]; !found {
				recordedChanges[unblockedChange] = struct{}{}
				blockedChanges.Unblock(unblockedChange)
				section = append(section, unblockedChange.Description())
			}
		}
		result = append(result, strings.Join(section, "\n"))

		if len(blocked) == 0 {
			return strings.Join(result, "\n---\n")
		}
	}
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
