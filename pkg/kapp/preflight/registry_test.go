// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package preflight

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diffgraph"
)

func TestRegistrySet(t *testing.T) {
	testCases := []struct {
		name       string
		preflights string
		registry   *Registry
		shouldErr  bool
	}{
		{
			name:       "no preflight checks registered, parsing skipped, any value can be provided",
			preflights: "someCheck",
			registry:   &Registry{},
		},
		{
			name:       "preflight checks registered, invalid check format in flag, error returned",
			preflights: ",",
			registry: &Registry{
				known: map[string]Check{
					"some": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph) error { return nil }, true),
				},
			},
			shouldErr: true,
		},
		{
			name:       "preflight checks registered, unknown preflight check specified, error returned",
			preflights: "nonexistent",
			registry: &Registry{
				known: map[string]Check{
					"exists": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph) error { return nil }, true),
				},
			},
			shouldErr: true,
		},
		{
			name:       "preflight checks registered, valid input, no error returned",
			preflights: "someCheck",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph) error { return nil }, true),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Set(tc.preflights)
			require.Equal(t, tc.shouldErr, err != nil)
		})
	}
}

func TestRegistryRun(t *testing.T) {
	testCases := []struct {
		name      string
		registry  *Registry
		shouldErr bool
	}{
		{
			name:     "no preflight checks registered, no error returned",
			registry: &Registry{},
		},
		{
			name: "preflight checks registered, disabled checks don't run",
			registry: &Registry{
				known: map[string]Check{
					"disabledCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph) error { return errors.New("should be disabled") }, false),
				},
			},
		},
		{
			name: "preflight checks registered, enabled check returns an error, error returned",
			registry: &Registry{
				known: map[string]Check{
					"errorCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph) error { return errors.New("error") }, true),
				},
			},
			shouldErr: true,
		},
		{
			name: "preflight checks registered, enabled checks successful, no error returned",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph) error { return nil }, true),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Run(nil, nil)
			require.Equal(t, tc.shouldErr, err != nil)
		})
	}
}
