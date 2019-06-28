package diffgraph

import (
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

var (
	crdsChangeGroup       = MustNewChangeGroupFromAnnString("change-groups.kapp.k14s.io/crds")
	namespacesChangeGroup = MustNewChangeGroupFromAnnString("change-groups.kapp.k14s.io/namespaces")
)

type ChangeDefaults struct {
	change ctldiff.Change
}

func (d ChangeDefaults) Groups() ([]ChangeGroup, error) {
	switch {
	case d.isCRD():
		return []ChangeGroup{crdsChangeGroup}, nil
	case d.isNs():
		return []ChangeGroup{namespacesChangeGroup}, nil
	default:
		return nil, nil
	}
}

func (d ChangeDefaults) AllRules() ([]ChangeRule, error) {
	// Not exact, but good enough match for ordering
	var rules []ChangeRule

	if !d.isCRD() {
		rules = append(rules, ChangeRule{
			Action:       ChangeRuleActionUpsert,
			Order:        ChangeRuleOrderAfter,
			TargetAction: ChangeRuleTargetActionUpserting,
			TargetGroup:  crdsChangeGroup,
		})

		// Do best effort to delete CRs before their CRDs
		rules = append(rules, ChangeRule{
			Action:       ChangeRuleActionDelete,
			Order:        ChangeRuleOrderBefore,
			TargetAction: ChangeRuleTargetActionDeleting,
			TargetGroup:  crdsChangeGroup,
		})
	}

	if len(d.change.NewOrExistingResource().Namespace()) > 0 {
		rules = append(rules, ChangeRule{
			Action:       ChangeRuleActionUpsert,
			Order:        ChangeRuleOrderAfter,
			TargetAction: ChangeRuleTargetActionUpserting,
			TargetGroup:  namespacesChangeGroup,
		})
	}

	return rules, nil
}

func (d ChangeDefaults) isCRD() bool {
	return ctlresm.NewApiExtensionsVxCRD(d.change.NewOrExistingResource()) != nil
}

func (d ChangeDefaults) isNs() bool {
	return ctlres.APIGroupKindMatcher{Kind: "Namespace"}.Matches(d.change.NewOrExistingResource())
}
