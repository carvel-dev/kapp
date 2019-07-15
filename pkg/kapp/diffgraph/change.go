package diffgraph

import (
	"fmt"
	"strings"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const (
	changeGroupAnnKey       = "kapp.k14s.io/change-group"
	changeGroupAnnPrefixKey = "kapp.k14s.io/change-group."

	changeRuleAnnKey       = "kapp.k14s.io/change-rule"
	changeRuleAnnPrefixKey = "kapp.k14s.io/change-rule."
)

type ActualChange interface {
	Resource() ctlres.Resource
	Op() ActualChangeOp
}

type ActualChangeOp string

const (
	ActualChangeOpUpsert ActualChangeOp = "upsert"
	ActualChangeOpDelete ActualChangeOp = "delete"
	ActualChangeOpNoop   ActualChangeOp = "noop"
)

type Change struct {
	Change     ActualChange
	WaitingFor []*Change

	groups *[]ChangeGroup
	rules  *[]ChangeRule
}

type Changes []*Change

func (c *Change) Groups() ([]ChangeGroup, error) {
	if c.groups != nil {
		return *c.groups, nil
	}

	var groups []ChangeGroup

	for k, v := range c.Change.Resource().Annotations() {
		if k == changeGroupAnnKey || strings.HasPrefix(k, changeGroupAnnPrefixKey) {
			groupKey, err := NewChangeGroupFromAnnString(v)
			if err != nil {
				return nil, err
			}
			groups = append(groups, groupKey)
		}
	}

	defaultGroups, err := ChangeDefaults{c.Change}.Groups()
	if err != nil {
		return nil, err
	}

	groups = append(groups, defaultGroups...)
	c.groups = &groups

	return groups, nil
}

func (c *Change) AllRules() ([]ChangeRule, error) {
	if c.rules != nil {
		return *c.rules, nil
	}

	var rules []ChangeRule

	for k, v := range c.Change.Resource().Annotations() {
		if k == changeRuleAnnKey || strings.HasPrefix(k, changeRuleAnnPrefixKey) {
			rule, err := NewChangeRuleFromAnnString(v)
			if err != nil {
				return nil, err
			}
			rules = append(rules, rule)
		}
	}

	defaultRules, err := ChangeDefaults{c.Change}.AllRules()
	if err != nil {
		return nil, err
	}

	rules = append(rules, defaultRules...)
	c.rules = &rules

	return rules, nil
}

func (c *Change) ApplicableRules() ([]ChangeRule, error) {
	var upsert, delete bool

	op := c.Change.Op()

	switch op {
	case ActualChangeOpUpsert:
		upsert = true
	case ActualChangeOpDelete:
		delete = true
	case ActualChangeOpNoop:
	default:
		return nil, fmt.Errorf("Unknown change operation: %s", op)
	}

	rules, err := c.AllRules()
	if err != nil {
		return nil, err
	}

	var applicableRules []ChangeRule
	for _, rule := range rules {
		if (upsert && rule.Action == ChangeRuleActionUpsert) ||
			(delete && rule.Action == ChangeRuleActionDelete) {
			applicableRules = append(applicableRules, rule)
		}
	}
	return applicableRules, nil
}

func (cs Changes) MatchesRule(rule ChangeRule, exceptChange *Change) ([]*Change, error) {
	var result []*Change

	for _, change := range cs {
		groups, err := change.Groups()
		if err != nil {
			return nil, err
		}

		for _, group := range groups {
			if !group.IsEqual(rule.TargetGroup) {
				continue
			}

			op := change.Change.Op()

			switch op {
			case ActualChangeOpUpsert:
				if rule.TargetAction == ChangeRuleTargetActionUpserting {
					result = append(result, change)
				}
			case ActualChangeOpDelete:
				if rule.TargetAction == ChangeRuleTargetActionDeleting {
					result = append(result, change)
				}
			case ActualChangeOpNoop:
			default:
				panic(fmt.Sprintf("Unknown change operation: %s", op))
			}
		}
	}

	return result, nil
}
