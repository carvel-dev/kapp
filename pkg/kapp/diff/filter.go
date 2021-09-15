// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"encoding/json"
	"fmt"

	"github.com/k14s/kapp/pkg/kapp/matcher"
	res "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/labels"
)

type ChangeSetFilter struct {
	df     DiffFilter
	Filter string
}

type DiffFilter struct {
	Kinds          []string
	Namespaces     []string
	Names          []string
	KindNames      []string
	KindNamespaces []string
	KindNsNames    []string
	LabelSelector  []string
	BoolFilter     *BoolFilter
}

type BoolFilter struct {
	And              []BoolFilter
	Or               []BoolFilter
	Not              *BoolFilter
	NewResource      *DiffFilter
	ExistingResource *DiffFilter
}

func (s *ChangeSetFilter) DiffFilter() (DiffFilter, error) {
	df := s.df
	if len(s.Filter) > 0 {
		boolFilter, err := NewBoolFilterFromString(s.Filter)
		if err != nil {
			return DiffFilter{}, err
		}
		df.BoolFilter = boolFilter
	}
	return df, nil
}

func NewBoolFilterFromString(data string) (*BoolFilter, error) {
	var filter BoolFilter

	err := json.Unmarshal([]byte(data), &filter)
	if err != nil {
		return nil, err
	}
	return &filter, nil
}

func (f DiffFilter) Apply(changes []Change) []Change {
	var result []Change

	for _, change := range changes {
		if f.Matches(change.NewResource(), change.ExistingResource()) {
			result = append(result, change)
		}
	}
	return result
}

func (f DiffFilter) Matches(newResource res.Resource, existingResource res.Resource) bool {
	if f.BoolFilter != nil && !f.BoolFilter.IsEmpty() {
		return f.BoolFilter.Matches(newResource, existingResource)
	}
	return true
}

func (f DiffFilter) MatchesNewResource(resource res.Resource) bool {

	if f.BoolFilter != nil {
		return f.BoolFilter.Matches(resource, nil)
	}

	return f.MatchesCommon(resource)
}

func (f DiffFilter) MatchesExistingResource(resource res.Resource) bool {

	if f.BoolFilter != nil {
		return f.BoolFilter.Matches(nil, resource)
	}

	return f.MatchesCommon(resource)
}

func (f DiffFilter) MatchesCommon(resource res.Resource) bool {
	if len(f.Kinds) > 0 {
		var matched bool
		for _, kind := range f.Kinds {
			if matcher.NewStringMatcher(kind).Matches(resource.Kind()) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.Namespaces) > 0 {
		var matched bool
		for _, ns := range f.Namespaces {
			if matcher.NewStringMatcher(ns).Matches(resource.Namespace()) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.Names) > 0 {
		var matched bool
		for _, name := range f.Names {
			if matcher.NewStringMatcher(name).Matches(resource.Name()) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
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

	if len(f.KindNames) > 0 {
		key := resource.Kind() + "/" + resource.Name()
		var matched bool
		for _, k := range f.KindNames {
			if key == k {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.KindNamespaces) > 0 {
		key := resource.Kind() + "/" + resource.Namespace()
		var matched bool
		for _, k := range f.KindNamespaces {
			if key == k {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(f.KindNsNames) > 0 {
		key := resource.Kind() + "/" + resource.Namespace() + "/" + resource.Name()
		var matched bool
		for _, k := range f.KindNsNames {
			if key == k {
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

func (m BoolFilter) Matches(newResource res.Resource, existingResource res.Resource) bool {
	if len(m.And) > 0 {
		for _, m2 := range m.And {
			if !m2.Matches(newResource, existingResource) {
				return false
			}
		}
		return true
	}

	if len(m.Or) > 0 {
		for _, m2 := range m.Or {
			if m2.Matches(newResource, existingResource) {
				return true
			}
		}
		return false
	}

	if m.Not != nil {
		return !m.Not.Matches(newResource, existingResource)
	}

	if m.NewResource != nil && newResource != nil {
		return m.NewResource.MatchesNewResource(newResource)
	}

	if m.ExistingResource != nil && existingResource != nil {
		return m.ExistingResource.MatchesExistingResource(existingResource)
	}
	return false
}

func (m BoolFilter) IsEmpty() bool {
	return m.And == nil && m.Or == nil && m.Not == nil && m.NewResource == nil && m.ExistingResource == nil
}
