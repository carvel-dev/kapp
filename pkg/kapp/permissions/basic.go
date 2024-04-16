// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package permissions

import (
	"context"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

// BasicValidator is a basic validator useful for
// validating basic CRUD permissions for resources. It has no knowledge
// of how to handle permission evaluation for specific
// GroupVersionKinds
type BasicValidator struct {
	ssarClient authv1client.SelfSubjectAccessReviewInterface
	mapper     meta.RESTMapper
}

var _ Validator = (*BasicValidator)(nil)

func NewBasicValidator(ssarClient authv1client.SelfSubjectAccessReviewInterface, mapper meta.RESTMapper) *BasicValidator {
	return &BasicValidator{
		ssarClient: ssarClient,
		mapper:     mapper,
	}
}

func (bv *BasicValidator) Validate(ctx context.Context, res ctlres.Resource, verb string) error {
	mapping, err := bv.mapper.RESTMapping(res.GroupKind(), res.GroupVersion().Version)
	if err != nil {
		return err
	}

	return ValidatePermissions(ctx, bv.ssarClient, &authv1.ResourceAttributes{
		Group:     mapping.Resource.Group,
		Version:   mapping.Resource.Version,
		Resource:  mapping.Resource.Resource,
		Namespace: res.Namespace(),
		Name:      res.Name(),
		Verb:      verb,
	})
}
