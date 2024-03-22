// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
	"strings"
)

type ChangeRuleAction string
type ChangeRuleOrder string
type ChangeRuleTargetAction string

const (
	ChangeRuleActionUpsert ChangeRuleAction = "upsert"
	ChangeRuleActionDelete ChangeRuleAction = "delete"

	ChangeRuleOrderBefore ChangeRuleOrder = "before"
	ChangeRuleOrderAfter  ChangeRuleOrder = "after"

	ChangeRuleTargetActionUpserting ChangeRuleTargetAction = "upserting"
	ChangeRuleTargetActionDeleting  ChangeRuleTargetAction = "deleting"
)

// Example: upsert before deleting apps.big.co/etcd
type ChangeRule struct {
	Action           ChangeRuleAction
	Order            ChangeRuleOrder
	TargetAction     ChangeRuleTargetAction
	TargetGroup      ChangeGroup
	IgnoreIfCyclical bool

	weight int
}

func NewChangeRuleFromAnnString(ann string) (ChangeRule, error) {
	pieces := strings.Split(ann, " ")
	if len(pieces) != 4 {
		return ChangeRule{}, fmt.Errorf(
			"Expected change rule annotation value to have format '(upsert|delete) (before|after) (upserting|deleting) (change-group)', but was '%s'", ann)
	}

	rule := ChangeRule{
		Action:       ChangeRuleAction(pieces[0]),
		Order:        ChangeRuleOrder(pieces[1]),
		TargetAction: ChangeRuleTargetAction(pieces[2]),
	}

	var err error

	rule.TargetGroup, err = NewChangeGroupFromAnnString(pieces[3])
	if err != nil {
		return ChangeRule{}, err
	}

	err = rule.Validate()
	if err != nil {
		return ChangeRule{}, err
	}

	return rule, nil
}

func (r ChangeRule) Validate() error {
	if r.Action != ChangeRuleActionUpsert && r.Action != ChangeRuleActionDelete {
		return fmt.Errorf("Unknown change rule Action")
	}
	if r.Order != ChangeRuleOrderBefore && r.Order != ChangeRuleOrderAfter {
		return fmt.Errorf("Unknown change rule Order")
	}
	if r.TargetAction != ChangeRuleTargetActionUpserting && r.TargetAction != ChangeRuleTargetActionDeleting {
		return fmt.Errorf("Unknown change rule TargetAction")
	}
	return nil
}
