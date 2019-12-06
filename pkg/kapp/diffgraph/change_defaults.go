package diffgraph

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

var (
	disableDefaultsAnnKey = "kapp.k14s.io/disable-default-change-group-and-rules"

	crdsChangeGroup       = MustNewChangeGroupFromAnnString("change-groups.kapp.k14s.io/crds")
	namespacesChangeGroup = MustNewChangeGroupFromAnnString("change-groups.kapp.k14s.io/namespaces")
)

type ChangeDefaults struct {
	change ActualChange
}

func (d ChangeDefaults) Groups() ([]ChangeGroup, error) {
	res := d.change.Resource()

	if annVal, found := res.Annotations()[disableDefaultsAnnKey]; found {
		if annVal != "" {
			return nil, fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
				disableDefaultsAnnKey, res.Description())
		}
		return nil, nil
	}

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
	res := d.change.Resource()

	if annVal, found := res.Annotations()[disableDefaultsAnnKey]; found {
		if annVal != "" {
			return nil, fmt.Errorf("Expected annotation '%s' on resource '%s' to have value ''",
				disableDefaultsAnnKey, res.Description())
		}
		return nil, nil
	}

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

	if len(res.Namespace()) > 0 {
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
	return ctlresm.NewApiExtensionsVxCRD(d.change.Resource()) != nil
}

func (d ChangeDefaults) isNs() bool {
	return ctlres.APIGroupKindMatcher{Kind: "Namespace"}.Matches(d.change.Resource())
}
