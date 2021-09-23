// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
	"strings"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
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
	ActualChangeOpExists ActualChangeOp = "exists"
)

type Change struct {
	Change     ActualChange
	WaitingFor []*Change

	changeGroupBindings []ctlconf.ChangeGroupBinding
	changeRuleBindings  []ctlconf.ChangeRuleBinding

	groups *[]ChangeGroup
	rules  *[]ChangeRule
}

type Changes []*Change

func (c *Change) Description() string {
	return fmt.Sprintf("(%s) %s", c.Change.Op(), c.Change.Resource().Description())
}

func (c *Change) IsDirectlyWaitingFor(changeToFind *Change) bool {
	for _, change := range c.WaitingFor {
		if change == changeToFind {
			return true
		}
	}
	return false
}

func (c *Change) IsTransitivelyWaitingFor(changeToFind *Change) bool {
	alreadyChecked := map[*Change]struct{}{}
	alreadyVisited := map[*Change]struct{}{}
	return c.isTransitivelyWaitingFor(changeToFind, alreadyChecked, alreadyVisited)
}

func (c *Change) isTransitivelyWaitingFor(changeToFind *Change,
	alreadyChecked map[*Change]struct{}, alreadyVisited map[*Change]struct{}) bool {

	if c.IsDirectlyWaitingFor(changeToFind) {
		return true
	}

	for _, change := range c.WaitingFor {
		if _, checked := alreadyChecked[change]; checked {
			continue
		}
		alreadyChecked[change] = struct{}{}

		// Should not happen, but let's double check to avoid infinite loops
		if _, visited := alreadyVisited[change]; visited {
			panic(fmt.Sprintf("Change: Internal error: cycle detected: %s",
				change.Change.Resource().Description()))
		}
		alreadyVisited[change] = struct{}{}

		if change.isTransitivelyWaitingFor(changeToFind, alreadyChecked, alreadyVisited) {
			return true
		}

		delete(alreadyVisited, change)
	}

	return false
}

func (c *Change) Groups() ([]ChangeGroup, error) {
	if c.groups != nil {
		return *c.groups, nil
	}

	var groups []ChangeGroup
	res := c.Change.Resource()

	for k, v := range res.Annotations() {
		if k == changeGroupAnnKey || strings.HasPrefix(k, changeGroupAnnPrefixKey) {
			name, err := NewChangeGroupNameForResource(v, c.Change.Resource()).AsString()
			if err != nil {
				return nil, err
			}
			groupKey, err := NewChangeGroupFromAnnString(name)
			if err != nil {
				return nil, err
			}
			groups = append(groups, groupKey)
		}
	}

	for _, groupConfig := range c.changeGroupBindings {
		rms := ctlconf.ResourceMatchers(groupConfig.ResourceMatchers).AsResourceMatchers()

		if (ctlres.AnyMatcher{rms}).Matches(res) {
			name, err := NewChangeGroupNameForResource(groupConfig.Name, c.Change.Resource()).AsString()
			if err != nil {
				return nil, err
			}
			groupKey, err := NewChangeGroupFromAnnString(name)
			if err != nil {
				return nil, err
			}
			groups = append(groups, groupKey)
		}
	}

	c.groups = &groups

	return groups, nil
}

func (c *Change) AllRules() ([]ChangeRule, error) {
	if c.rules != nil {
		return *c.rules, nil
	}

	var rules []ChangeRule
	res := c.Change.Resource()

	for k, v := range res.Annotations() {
		if k == changeRuleAnnKey || strings.HasPrefix(k, changeRuleAnnPrefixKey) {
			ruleStr, err := NewChangeGroupNameForResource(v, c.Change.Resource()).AsString()
			if err != nil {
				return nil, err
			}
			rule, err := NewChangeRuleFromAnnString(ruleStr)
			if err != nil {
				return nil, fmt.Errorf("Resource %s: %s", res.Description(), err)
			}
			rules = append(rules, rule)
		}
	}

	for i, ruleConfig := range c.changeRuleBindings {
		rms := ctlconf.ResourceMatchers(ruleConfig.ResourceMatchers).AsResourceMatchers()

		if (ctlres.AnyMatcher{rms}).Matches(res) {
			for _, ruleStr := range ruleConfig.Rules {
				ruleStr, err := NewChangeGroupNameForResource(ruleStr, c.Change.Resource()).AsString()
				if err != nil {
					return nil, err
				}
				rule, err := NewChangeRuleFromAnnString(ruleStr)
				if err != nil {
					return nil, fmt.Errorf("Resource %s: %s", res.Description(), err)
				}
				rule.IgnoreIfCyclical = ruleConfig.IgnoreIfCyclical
				rule.weight = 100 + i // start at 100
				rules = append(rules, rule)
			}
		}
	}

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
	case ActualChangeOpExists:
		upsert = true
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
			case ActualChangeOpUpsert, ActualChangeOpExists:
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
