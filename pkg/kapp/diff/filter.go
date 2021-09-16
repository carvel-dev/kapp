// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"encoding/json"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeSetFilter struct {
	Filter string
}

type OpsFilter []ChangeOp

type ChangeSetFilterRoot struct {
	And              []ChangeSetFilterRoot
	Or               []ChangeSetFilterRoot
	Not              *ChangeSetFilterRoot
	Ops              OpsFilter
	NewResource      *ctlres.ResourceFilter
	ExistingResource *ctlres.ResourceFilter
}

func (s *ChangeSetFilter) DiffFilter() (ChangeSetFilterRoot, error) {
	var filter ChangeSetFilterRoot
	if len(s.Filter) > 0 {
		changeSetFilter, err := NewChangeSetFilterFromString(s.Filter)
		if err != nil {
			return ChangeSetFilterRoot{}, err
		}
		filter = *changeSetFilter
	}
	return filter, nil
}

func NewChangeSetFilterFromString(data string) (*ChangeSetFilterRoot, error) {
	var filter ChangeSetFilterRoot

	err := json.Unmarshal([]byte(data), &filter)
	if err != nil {
		return nil, err
	}
	return &filter, nil
}

func (f ChangeSetFilterRoot) Apply(changes []Change) []Change {
	if f.IsEmpty() {
		return changes
	}
	var result []Change

	for _, change := range changes {
		if f.Matches(change) {
			result = append(result, change)
		}
	}
	return result
}

func (ops OpsFilter) Matches(change Change) bool {
	for _, op := range ops {
		if op != change.Op() {
			return false
		}
	}
	return true
}

func (m ChangeSetFilterRoot) Matches(change Change) bool {
	if len(m.And) > 0 {
		for _, m2 := range m.And {
			if !m2.Matches(change) {
				return false
			}
		}
		return true
	}

	if len(m.Or) > 0 {
		for _, m2 := range m.Or {
			if m2.Matches(change) {
				return true
			}
		}
		return false
	}

	if m.Not != nil {
		return !m.Not.Matches(change)
	}

	if m.NewResource != nil && change.NewResource() != nil {
		return m.NewResource.Matches(change.NewResource())
	}

	if m.ExistingResource != nil && change.ExistingResource() != nil {
		return m.ExistingResource.Matches(change.ExistingResource())
	}

	if len(m.Ops) > 0 && change.Op() != "" {
		return m.Ops.Matches(change)
	}
	return false
}

func (m ChangeSetFilterRoot) IsEmpty() bool {
	return m.And == nil && m.Or == nil && m.Not == nil && m.NewResource == nil && m.ExistingResource == nil && m.Ops == nil
}
