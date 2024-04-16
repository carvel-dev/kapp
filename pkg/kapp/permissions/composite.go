// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package permissions

import (
	"context"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ Validator = (*CompositeValidator)(nil)

// CompositeValidator implements Validator and is used
// for composing multiple validators into a single validator
// that can handle specifying unique validators for different
// GroupVersionKinds
type CompositeValidator struct {
	validators       map[schema.GroupVersionKind]Validator
	defaultValidator Validator
}

func NewCompositeValidator(defaultValidator Validator, validators map[schema.GroupVersionKind]Validator) *CompositeValidator {
	return &CompositeValidator{
		validators:       validators,
		defaultValidator: defaultValidator,
	}
}

func (cv *CompositeValidator) Validate(ctx context.Context, res ctlres.Resource, verb string) error {
	if validator, ok := cv.validators[res.GroupVersion().WithKind(res.Kind())]; ok {
		return validator.Validate(ctx, res, verb)
	}
	return cv.defaultValidator.Validate(ctx, res, verb)
}
