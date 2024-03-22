// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package permissions

import (
	"context"
	"errors"
	"fmt"

	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
	authv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/component-helpers/auth/rbac/validation"
)

// RoleValidator is a Validator implementation
// for validating permissions required to CRUD
// Kubernetes (Cluster)Role resources
type RoleValidator struct {
	ssarClient authv1client.SelfSubjectAccessReviewInterface
	mapper     meta.RESTMapper
}

var _ Validator = (*RoleValidator)(nil)

func NewRoleValidator(ssarClient authv1client.SelfSubjectAccessReviewInterface, mapper meta.RESTMapper) *RoleValidator {
	return &RoleValidator{
		ssarClient: ssarClient,
		mapper:     mapper,
	}
}

func (rv *RoleValidator) Validate(ctx context.Context, res ctlres.Resource, verb string) error {
	mapping, err := rv.mapper.RESTMapping(res.GroupKind(), res.GroupVersion().Version)
	if err != nil {
		return err
	}

	switch verb {
	case "create", "update":
		// do early validation on create / update to see if a user has
		// the "escalate" permissions which allows them to perform
		// privilege escalation and create any (Cluster)Role
		err := ValidatePermissions(ctx, rv.ssarClient, &authv1.ResourceAttributes{
			Group:     mapping.Resource.Group,
			Version:   mapping.Resource.Version,
			Resource:  mapping.Resource.Resource,
			Namespace: res.Namespace(),
			Name:      res.Name(),
			Verb:      "escalate",
		})
		// if the error is nil, the user has the "escalate" permissions so we should
		// return early. Otherwise, they don't have the "escalate" permissions and
		// we need to continue our validations.
		if err == nil {
			return nil
		}

		// Check if user has permissions to even create/update the resource
		err = ValidatePermissions(ctx, rv.ssarClient, &authv1.ResourceAttributes{
			Group:     mapping.Resource.Group,
			Version:   mapping.Resource.Version,
			Resource:  mapping.Resource.Resource,
			Namespace: res.Namespace(),
			Name:      res.Name(),
			Verb:      verb,
		})
		if err != nil {
			return err
		}
		// If user doesn't have "escalate" permissions then they can
		// only create (Cluster)Roles that contain permissions they already have.
		// Loop through all the defined policies and determine
		// if a user has the appropriate permissions
		rules, err := RulesForRole(res)
		if err != nil {
			return fmt.Errorf("parsing rules for role: %w", err)
		}

		errorSet := []error{}
		for _, rule := range rules {
			// breakdown the rules into the subset of
			// rules such that the subrules contain
			// at most one verb, one group, and one resource
			// source at: https://github.com/kubernetes/component-helpers/blob/9a5801419916272fc9cec7a7822ed525721b99d3/auth/rbac/validation/policy_comparator.go#L56-L84
			var subrules []rbacv1.PolicyRule = validation.BreakdownRule(rule)
			for _, subrule := range subrules {
				resourceName := ""
				if len(subrule.ResourceNames) > 0 {
					resourceName = subrule.ResourceNames[0]
				}
				err := ValidatePermissions(ctx, rv.ssarClient, &authv1.ResourceAttributes{
					Group:     subrule.APIGroups[0],
					Resource:  subrule.Resources[0],
					Namespace: res.Namespace(),
					Name:      resourceName,
					Verb:      subrule.Verbs[0],
				})
				if err != nil {
					errorSet = append(errorSet, err)
				}
			}
		}

		if len(errorSet) > 0 {
			baseErr := fmt.Errorf("potential privilege escalation, not permitted to %q %s", verb, res.GroupVersion().WithKind(res.Kind()).String())
			return errors.Join(append([]error{baseErr}, errorSet...)...)
		}
	default:
		return ValidatePermissions(ctx, rv.ssarClient, &authv1.ResourceAttributes{
			Group:     mapping.Resource.Group,
			Version:   mapping.Resource.Version,
			Resource:  mapping.Resource.Resource,
			Namespace: res.Namespace(),
			Name:      res.Name(),
			Verb:      verb,
		})
	}

	return nil
}
