// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"encoding/json"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
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

func (s *ChangeSetFilter) DiffFilter() (*ChangeSetFilterRoot, error) {
	if len(s.Filter) == 0 {
		return &ChangeSetFilterRoot{}, nil
	}
	return NewChangeSetFilterRootFromString(s.Filter)
}

func NewChangeSetFilterRootFromString(data string) (*ChangeSetFilterRoot, error) {
	var filter ChangeSetFilterRoot

	err := json.Unmarshal([]byte(data), &filter)
	if err != nil {
		return nil, err
	}
	return &filter, nil
}

func (f ChangeSetFilterRoot) Apply(changes []Change) []Change {
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

	// case when user is trying to filter on existing resource but existing resource not exists
	if f.ExistingResource != nil && change.ExistingResource() == nil {
		return false
	}

	// case when user is trying to filter on new resource but new resource not exists
	if f.NewResource != nil && change.NewResource() == nil {
		return false
	}

	if len(f.Ops) > 0 && change.Op() != "" {
		return f.Ops.Matches(change)
	}
	return true
}
