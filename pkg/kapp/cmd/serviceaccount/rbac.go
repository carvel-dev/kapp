// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package serviceaccount

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"sigs.k8s.io/yaml"
)

type ServiceAccount struct {
	resource ctlres.Resource
	used     bool
}

func NewServiceAccount(resource ctlres.Resource) (*ServiceAccount, error) {
	return &ServiceAccount{resource, false}, nil
}

func (s *ServiceAccount) APIGroup() string  { return s.resource.APIGroup() }
func (s *ServiceAccount) Kind() string      { return s.resource.Kind() }
func (s *ServiceAccount) Name() string      { return s.resource.Name() }
func (s *ServiceAccount) Namespace() string { return s.resource.Namespace() }

func (s *ServiceAccount) MarkUsed()  { s.used = true }
func (s *ServiceAccount) Used() bool { return s.used }

type Role struct {
	resource ctlres.Resource
	schema   schemaRole
	used     bool
}

func NewRole(resource ctlres.Resource) (*Role, error) {
	role := &Role{resource: resource}

	bs, err := resource.AsYAMLBytes()
	if err != nil {
		return role, err
	}

	err = yaml.Unmarshal(bs, &role.schema)
	if err != nil {
		return role, err
	}

	return role, nil
}

func (r *Role) APIGroup() string  { return r.resource.APIGroup() }
func (r *Role) Kind() string      { return r.resource.Kind() }
func (r *Role) Name() string      { return r.resource.Name() }
func (r *Role) Namespace() string { return r.resource.Namespace() }

func (r *Role) MarkUsed()  { r.used = true }
func (r *Role) Used() bool { return r.used }

func (r *Role) Verbs() []string {
	return r.uniqueStrings(func(r schemaRoleRule) []string { return r.Verbs })
}

func (r *Role) APIGroups() []string {
	return r.uniqueStrings(func(r schemaRoleRule) []string { return r.APIGroups })
}

func (r *Role) Resources() []string {
	return r.uniqueStrings(func(r schemaRoleRule) []string { return r.Resources })
}

func (r *Role) uniqueStrings(f func(rule schemaRoleRule) []string) []string {
	var result []string
	unique := map[string]struct{}{}
	for _, rule := range r.schema.Rules {
		for _, str := range f(rule) {
			if _, ok := unique[str]; !ok {
				result = append(result, str)
				unique[str] = struct{}{}
			}
		}
	}
	return result
}

type RoleBinding struct {
	resource ctlres.Resource
	schema   schemaRoleBinding
	used     bool
}

func NewRoleBinding(resource ctlres.Resource) (*RoleBinding, error) {
	binding := &RoleBinding{resource: resource}

	bs, err := resource.AsYAMLBytes()
	if err != nil {
		return binding, err
	}

	err = yaml.Unmarshal(bs, &binding.schema)
	if err != nil {
		return binding, err
	}

	return binding, nil
}

func (r *RoleBinding) Name() string      { return r.resource.Name() }
func (r *RoleBinding) Namespace() string { return r.resource.Namespace() }

func (r *RoleBinding) MarkUsed()  { r.used = true }
func (r *RoleBinding) Used() bool { return r.used }

func (r *RoleBinding) MatchesRole(role *Role) bool {
	if r.schema.RoleRef.Name == role.Name() &&
		r.schema.RoleRef.Kind == role.Kind() &&
		r.schema.RoleRef.APIGroup == role.APIGroup() &&
		r.resource.Namespace() == role.Namespace() {
		return true
	}
	return false
}

func (r *RoleBinding) MatchesServiceAccount(sa *ServiceAccount) bool {
	for _, subj := range r.schema.Subjects {
		if subj.Name == sa.Name() &&
			subj.Kind == sa.Kind() &&
			subj.APIGroup == sa.APIGroup() &&
			subj.Namespace == sa.Namespace() {
			return true
		}
	}
	return false
}

type schemaRoleBinding struct {
	RoleRef  schemaRoleRef
	Subjects []schemaSubject
}

type schemaRoleRef struct {
	Name     string
	Kind     string
	APIGroup string
}

type schemaSubject struct {
	Name      string
	Kind      string
	APIGroup  string
	Namespace string
}

type schemaRole struct {
	Rules []schemaRoleRule
}

type schemaRoleRule struct {
	APIGroups []string `yaml:"apiGroups"`
	Resources []string
	Verbs     []string
}
