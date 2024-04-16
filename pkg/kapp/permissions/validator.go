// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package permissions

import (
	"context"
	"errors"
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	authv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	rbacv1client "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

type Validator interface {
	Validate(context.Context, ctlres.Resource, string) error
}

// ValidatePermissons takes in all the parameters necessary to validate permissions using a
// SelfSubjectAccessReview. It returns an error if the SelfSubjectAccessReview indicates that
// the permissions are not present or are unable to be determined. A nil error is returned if
// the SelfSubjectAccessReview indicates that the permissions are present.
// TODO: Look into using SelfSubjectRulesReview instead of SelfSubjectAccessReview
func ValidatePermissions(ctx context.Context, ssarClient authv1client.SelfSubjectAccessReviewInterface, resourceAttributes *authv1.ResourceAttributes) error {
	ssar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: resourceAttributes,
		},
	}

	retSsar, err := ssarClient.Create(ctx, ssar, v1.CreateOptions{})
	if err != nil {
		return err
	}

	if retSsar == nil {
		return errors.New("unable to validate permissions: returned SelfSubjectAccessReview is nil")
	}

	if retSsar.Status.EvaluationError != "" {
		return fmt.Errorf("unable to validate permissions: %s", retSsar.Status.EvaluationError)
	}

	if !retSsar.Status.Allowed {
		gvr := schema.GroupVersionResource{
			Group:    resourceAttributes.Group,
			Version:  resourceAttributes.Version,
			Resource: resourceAttributes.Resource,
		}
		return fmt.Errorf("not permitted to %q %s",
			resourceAttributes.Verb,
			gvr.String())
	}

	return nil
}

// RulesForRole will return a slice of rbacv1.PolicyRule objects
// that are representative of a provided (Cluster)Role's rules.
// It returns an error if one occurs during the process of fetching this
// information or if it is unable to determine the kind of binding this is
func RulesForRole(res ctlres.Resource) ([]rbacv1.PolicyRule, error) {
	switch res.Kind() {
	case "Role":
		role := &rbacv1.Role{}
		err := res.AsTypedObj(role)
		if err != nil {
			return nil, fmt.Errorf("converting resource to typed Role object: %w", err)
		}

		return role.Rules, nil

	case "ClusterRole":
		role := &rbacv1.ClusterRole{}
		err := res.AsTypedObj(role)
		if err != nil {
			return nil, fmt.Errorf("converting resource to typed ClusterRole object: %w", err)
		}

		return role.Rules, nil
	}

	return nil, fmt.Errorf("unknown role kind %q", res.Kind())
}

// RulesForBinding will return a slice of rbacv1.PolicyRule objects
// that are representative of the (Cluster)Role rules that a (Cluster)RoleBinding
// references. It returns an error if one occurs during the process of fetching this
// information or if it is unable to determine the kind of binding this is
func RulesForBinding(ctx context.Context, rbacClient rbacv1client.RbacV1Interface, res ctlres.Resource) ([]rbacv1.PolicyRule, error) {
	switch res.Kind() {
	case "RoleBinding":
		roleBinding := &rbacv1.RoleBinding{}
		err := res.AsTypedObj(roleBinding)
		if err != nil {
			return nil, fmt.Errorf("converting resource to typed RoleBinding object: %w", err)
		}

		return RulesForRoleBinding(ctx, rbacClient, roleBinding)
	case "ClusterRoleBinding":
		roleBinding := &rbacv1.ClusterRoleBinding{}
		err := res.AsTypedObj(roleBinding)
		if err != nil {
			return nil, fmt.Errorf("converting resource to typed ClusterRoleBinding object: %w", err)
		}

		return RulesForClusterRoleBinding(ctx, rbacClient, roleBinding)
	}

	return nil, fmt.Errorf("unknown binding kind %q", res.Kind())
}

// RulesForRoleBinding will return a slice of rbacv1.PolicyRule objects
// that are representative of the (Cluster)Role rules that a RoleBinding
// references. It returns an error if one occurs during the process of fetching this
// information.
func RulesForRoleBinding(ctx context.Context, rbacClient rbacv1client.RbacV1Interface, rb *rbacv1.RoleBinding) ([]rbacv1.PolicyRule, error) {
	switch rb.RoleRef.Kind {
	case "ClusterRole":
		role, err := rbacClient.ClusterRoles().Get(ctx, rb.RoleRef.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("fetching ClusterRole %q for RoleBinding %q: %w", rb.RoleRef.Name, rb.Name, err)
		}

		return role.Rules, nil
	case "Role":
		role, err := rbacClient.Roles(rb.Namespace).Get(ctx, rb.RoleRef.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("fetching Role %q for RoleBinding %q: %w", rb.RoleRef.Name, rb.Name, err)
		}

		return role.Rules, nil
	}

	return nil, fmt.Errorf("unknown role reference kind: %q", rb.RoleRef.Kind)
}

// RulesForClusterRoleBinding will return a slice of rbacv1.PolicyRule objects
// that are representative of the ClusterRole rules that a ClusterRoleBinding
// references. It returns an error if one occurs during the process of fetching this
// information.
func RulesForClusterRoleBinding(ctx context.Context, crGetter rbacv1client.ClusterRolesGetter, crb *rbacv1.ClusterRoleBinding) ([]rbacv1.PolicyRule, error) {
	role, err := crGetter.ClusterRoles().Get(ctx, crb.RoleRef.Name, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("fetching ClusterRole %q for ClusterRoleBinding %q: %w", crb.RoleRef.Name, crb.Name, err)
	}

	return role.Rules, nil
}
