// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package crdupgradesafety

import (
	"context"
	"errors"
	"fmt"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	ctldgraph "carvel.dev/kapp/pkg/kapp/diffgraph"
	"carvel.dev/kapp/pkg/kapp/preflight"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ preflight.Check = (*Preflight)(nil)

// Preflight is an implementation of preflight.Check
// to make it easier to add crd upgrade validation
// as a preflight check
type Preflight struct {
	depsFactory cmdcore.DepsFactory
	enabled     bool
	validator   *Validator
}

func NewPreflight(df cmdcore.DepsFactory, enabled bool) *Preflight {
	return &Preflight{
		depsFactory: df,
		enabled:     enabled,
		validator: &Validator{
			Validations: []Validation{
				NewValidationFunc("NoScopeChange", NoScopeChange),
				NewValidationFunc("NoStoredVersionRemoved", NoStoredVersionRemoved),
				NewValidationFunc("NoExistingFieldRemoved", NoExistingFieldRemoved),
				&ChangeValidator{
					Validations: []ChangeValidation{
						EnumChangeValidation,
					},
				},
			},
		},
	}
}

func (p *Preflight) Enabled() bool {
	return p.enabled
}

func (p *Preflight) SetEnabled(enabled bool) {
	p.enabled = enabled
}

func (p *Preflight) SetConfig(_ preflight.CheckConfig) error {
	return nil
}

func (p *Preflight) Run(ctx context.Context, changeGraph *ctldgraph.ChangeGraph) error {
	dCli, err := p.depsFactory.DynamicClient(cmdcore.DynamicClientOpts{})
	if err != nil {
		return fmt.Errorf("getting dynamic client: %w", err)
	}
	crdCli := dCli.Resource(v1.SchemeGroupVersion.WithResource("customresourcedefinitions"))

	validateErrs := []error{}
	for _, change := range changeGraph.All() {
		// Loop through all the changes looking for "upsert" operations on
		// a CRD. "upsert" is used for create + update operations
		if change.Change.Op() != ctldgraph.ActualChangeOpUpsert {
			continue
		}
		res := change.Change.Resource()
		if res.GroupVersion().WithKind(res.Kind()) != v1.SchemeGroupVersion.WithKind("CustomResourceDefinition") {
			continue
		}

		// to properly determine if this is an update operation, attempt to fetch
		// the "old" CRD from the cluster
		uOldCRD, err := crdCli.Get(ctx, res.Name(), metav1.GetOptions{})
		if err != nil {
			// if the resource is not found, this "upsert" operation
			// translates to a "create" request being made. Skip this change
			if apierrors.IsNotFound(err) {
				continue
			}

			return fmt.Errorf("checking for existing CRD resource: %w", err)
		}

		oldCRD := &v1.CustomResourceDefinition{}
		s := runtime.NewScheme()
		if err := v1.AddToScheme(s); err != nil {
			return fmt.Errorf("adding apiextension apis to scheme: %w", err)
		}
		if err := s.Convert(uOldCRD, oldCRD, nil); err != nil {
			return fmt.Errorf("couldn't convert old CRD resource to a CRD object: %w", err)
		}

		newCRD := &v1.CustomResourceDefinition{}
		if err := res.AsUncheckedTypedObj(newCRD); err != nil {
			return fmt.Errorf("couldn't convert new CRD resource to a CRD object: %w", err)
		}

		if err = p.validator.Validate(*oldCRD, *newCRD); err != nil {
			validateErrs = append(validateErrs, err)
		}
	}

	if len(validateErrs) > 0 {
		baseErr := errors.New("validation for safe CRD upgrades failed")
		return errors.Join(append([]error{baseErr}, validateErrs...)...)
	}

	return nil
}
