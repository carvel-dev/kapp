package diffgraph

import (
	"fmt"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
)

type ChangeGraph struct {
	changes []*Change
}

func NewChangeGraph(changes []ctldiff.Change) (*ChangeGraph, error) {
	graphChanges := []*Change{}

	for _, change := range changes {
		graphChanges = append(graphChanges, &Change{Change: change})
	}

	for _, graphChange := range graphChanges {
		rules, err := graphChange.ApplicableRules()
		if err != nil {
			return nil, err
		}

		for _, rule := range rules {
			switch {
			case rule.Order == ChangeRuleOrderAfter:
				matchedChanges, err := Changes(graphChanges).MatchesRule(rule, graphChange)
				if err != nil {
					return nil, err
				}
				graphChange.WaitingFor = append(graphChange.WaitingFor, matchedChanges...)

			case rule.Order == ChangeRuleOrderBefore:
				matchedChanges, err := Changes(graphChanges).MatchesRule(rule, graphChange)
				if err != nil {
					return nil, err
				}
				for _, matchedChange := range matchedChanges {
					matchedChange.WaitingFor = append(matchedChange.WaitingFor, graphChange)
				}
			}
		}
	}

	graph := &ChangeGraph{graphChanges}

	return graph, graph.checkCycles()
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

func (g *ChangeGraph) Print() {
	fmt.Printf("%s", g.PrintStr())
}

func (g *ChangeGraph) PrintStr() string {
	return g.printChanges(g.changes, "")
}

func (g *ChangeGraph) printChanges(changes []*Change, indent string) string {
	var result string
	for _, change := range changes {
		result += fmt.Sprintf("%s(%s) %s\n", indent, change.Change.Op(), change.Change.NewOrExistingResource().Description())
		result += g.printChanges(change.WaitingFor, indent+"  ")
	}
	return result
}

func (g *ChangeGraph) checkCycles() error {
	for _, change := range g.changes {
		changeDesc := change.Change.NewOrExistingResource().Description()

		err := g.checkCyclesInChanges(change.WaitingFor, change, changeDesc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *ChangeGraph) checkCyclesInChanges(changes []*Change, selfChange *Change, descPath string) error {
	for _, change := range changes {
		changeDesc := fmt.Sprintf("%s -> %s", descPath, change.Change.NewOrExistingResource().Description())
		if change == selfChange {
			return fmt.Errorf("Detected cycle in grouped changes: %s", changeDesc)
		}

		err := g.checkCyclesInChanges(change.WaitingFor, selfChange, changeDesc)
		if err != nil {
			return err
		}
	}
	return nil
}
