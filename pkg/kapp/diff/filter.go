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

func (f ChangeSetFilterRoot) Matches(change Change) bool {
	if len(f.And) > 0 {
		for _, m2 := range f.And {
			if !m2.Matches(change) {
				return false
			}
		}
		return true
	}

	if len(f.Or) > 0 {
		for _, m2 := range f.Or {
			if m2.Matches(change) {
				return true
			}
		}
		return false
	}

	if f.Not != nil {
		return !f.Not.Matches(change)
	}

	if f.NewResource != nil && change.NewResource() != nil {
		return f.NewResource.Matches(change.NewResource())
	}

	if f.ExistingResource != nil && change.ExistingResource() != nil {
		return f.ExistingResource.Matches(change.ExistingResource())
	}

	if len(f.Ops) > 0 && change.Op() != "" {
		return f.Ops.Matches(change)
	}
	return false
}

func (f ChangeSetFilterRoot) IsEmpty() bool {
	return f.And == nil && f.Or == nil && f.Not == nil && f.NewResource == nil && f.ExistingResource == nil && f.Ops == nil
}
