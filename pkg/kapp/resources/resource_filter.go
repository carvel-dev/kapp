// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/k14s/kapp/pkg/kapp/matcher" // TODO inject
	"k8s.io/apimachinery/pkg/labels"
)

type ResourceFilter struct {
	CreatedAtBeforeTime *time.Time
	CreatedAtAfterTime  *time.Time

	Kinds          []string
	Namespaces     []string
	Names          []string
	KindNames      []string
	KindNamespaces []string
	KindNsNames    []string
	Labels         []string

	BoolFilter *BoolFilter `json:"-"`
}

func (f ResourceFilter) Apply(resources []Resource) []Resource {
	var result []Resource

	for _, resource := range resources {
		if f.Matches(resource) {
			result = append(result, resource)
		}
	}

	return result
}

func (f ResourceFilter) Matches(resource Resource) bool {
	if f.BoolFilter != nil {
		return f.BoolFilter.Matches(resource)
	}

	if f.CreatedAtBeforeTime != nil {
		if resource.CreatedAt().After(*f.CreatedAtBeforeTime) {
			return false
		}
	}

	if f.CreatedAtAfterTime != nil {
		if resource.CreatedAt().Before(*f.CreatedAtAfterTime) {
			return false
		}
	}

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

	if len(f.Labels) > 0 {
		var matched bool
		for _, label := range f.Labels {
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

type BoolFilter struct {
	And      []BoolFilter
	Or       []BoolFilter
	Not      *BoolFilter
	Resource *ResourceFilter
}

func NewBoolFilterFromString(data string) (*BoolFilter, error) {
	var filter BoolFilter

	err := json.Unmarshal([]byte(data), &filter)
	if err != nil {
		return nil, err
	}

	return &filter, nil
}

func (m BoolFilter) Matches(res Resource) bool {
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

	if m.Resource != nil {
		return m.Resource.Matches(res)
	}

	return false
}
