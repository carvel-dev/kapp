package diffgraph

import (
	"fmt"
	"strings"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
)

const (
	changeGroupAnnKey       = "kapp.k14s.io/change-group"
	changeGroupAnnPrefixKey = "kapp.k14s.io/change-group."

	changeRuleAnnKey       = "kapp.k14s.io/change-rule"
	changeRuleAnnPrefixKey = "kapp.k14s.io/change-rule."
)

type Change struct {
	Change     ctldiff.Change
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

	for k, v := range c.Change.NewOrExistingResource().Annotations() {
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

	for k, v := range c.Change.NewOrExistingResource().Annotations() {
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
	op := c.Change.Op()
	var upsert, delete bool

	switch op {
	case ctldiff.ChangeOpAdd, ctldiff.ChangeOpUpdate:
		upsert = true
	case ctldiff.ChangeOpDelete:
		delete = true
	case ctldiff.ChangeOpKeep:
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
			case ctldiff.ChangeOpAdd, ctldiff.ChangeOpUpdate:
				if rule.TargetAction == ChangeRuleTargetActionUpserting {
					result = append(result, change)
				}
			case ctldiff.ChangeOpDelete:
				if rule.TargetAction == ChangeRuleTargetActionDeleting {
					result = append(result, change)
				}
			case ctldiff.ChangeOpKeep:
			default:
				panic(fmt.Sprintf("Unknown change operation: %s", op))
			}
		}
	}

	return result, nil
}
