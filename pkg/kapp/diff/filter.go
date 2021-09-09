// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"encoding/json"
	"fmt"

	res "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/labels"
)

type ChangeSetFilter struct {
	lf     DiffFilter
	Filter string
}

type DiffFilter struct {
	LabelSelector []string
	BoolFilter    *BoolFilter
}

type BoolFilter struct {
	And              []BoolFilter
	Or               []BoolFilter
	Not              *BoolFilter
	NewResource      *DiffFilter
	ExistingResource *DiffFilter
}

func (s *ChangeSetFilter) DiffFilter() (DiffFilter, error) {
	lf := s.lf
	if len(s.Filter) > 0 {
		boolFilter, err := NewBoolFilterFromString(s.Filter)
		if err != nil {
			return DiffFilter{}, err
		}

		lf.BoolFilter = boolFilter
	}
	return lf, nil
}

func NewBoolFilterFromString(data string) (*BoolFilter, error) {
	var filter BoolFilter

	err := json.Unmarshal([]byte(data), &filter)
	if err != nil {
		return nil, err
	}

	return &filter, nil
}

func (f DiffFilter) Apply(existingResources []res.Resource) []res.Resource {
	var result []res.Resource

	for _, resource := range existingResources {

		if f.Matches(resource) {
			result = append(result, resource)
		}
	}

	return result
}

func (f DiffFilter) Matches(resource res.Resource) bool {
	if f.BoolFilter != nil {
		return f.BoolFilter.Matches(resource)
	}

	if len(f.LabelSelector) > 0 {
		var matched bool
		for _, label := range f.LabelSelector {
			labelSelector, err := labels.Parse(label)
			if err != nil {
				panic(fmt.Sprintf("Parsing label selector failed: %s", err))
			}
			if labelSelector.Matches(labels.Set(resource.Labels())) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (m BoolFilter) Matches(res res.Resource) bool {
	if len(m.And) > 0 {
		for _, m2 := range m.And {
			if !m2.Matches(res) {
				return false
			}
		}
		return true
	}

	if len(m.Or) > 0 {
		for _, m2 := range m.Or {
			if m2.Matches(res) {
				return true
			}
		}
		return false
	}

	if m.Not != nil {
		return !m.Not.Matches(res)
	}

	if m.NewResource != nil {
		return m.NewResource.Matches(res)
	}

	if m.ExistingResource != nil {
		return m.ExistingResource.Matches(res)
	}

	return false
}
