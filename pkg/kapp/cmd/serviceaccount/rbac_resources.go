// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package serviceaccount

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type RBACResources struct {
	ServiceAccounts []*ServiceAccount
	Roles           []*Role
	RoleBindings    []*RoleBinding
}

func (r *RBACResources) Collect(resources []ctlres.Resource) error {
	serviceAccountMatcher := ctlres.AnyMatcher{[]ctlres.ResourceMatcher{
		ctlres.APIVersionKindMatcher{APIVersion: "v1", Kind: "ServiceAccount"},
	}}

	roleMatcher := ctlres.AnyMatcher{[]ctlres.ResourceMatcher{
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1beta1", Kind: "ClusterRole"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1beta1", Kind: "Role"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1alpha1", Kind: "ClusterRole"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1alpha1", Kind: "Role"},
	}}

	roleBindingMatcher := ctlres.AnyMatcher{[]ctlres.ResourceMatcher{
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRoleBinding"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1beta1", Kind: "ClusterRoleBinding"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1beta1", Kind: "RoleBinding"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1alpha1", Kind: "ClusterRoleBinding"},
		ctlres.APIVersionKindMatcher{APIVersion: "rbac.authorization.k8s.io/v1alpha1", Kind: "RoleBinding"},
	}}

	for _, res := range resources {
		switch {
		case serviceAccountMatcher.Matches(res):
			sa, err := NewServiceAccount(res)
			if err != nil {
				return err
			}
			r.ServiceAccounts = append(r.ServiceAccounts, sa)

		case roleMatcher.Matches(res):
			role, err := NewRole(res)
			if err != nil {
				return err
			}
			r.Roles = append(r.Roles, role)

		case roleBindingMatcher.Matches(res):
			roleBinding, err := NewRoleBinding(res)
			if err != nil {
				return err
			}
			r.RoleBindings = append(r.RoleBindings, roleBinding)
		}
	}

	return nil
}
