// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0
package preflight

import (
	"context"
	"errors"
	"testing"

	ctlconf "carvel.dev/kapp/pkg/kapp/config"
	"carvel.dev/kapp/pkg/kapp/diffgraph"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
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
					"some": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
				},
				enabledFlag: map[string]bool{},
			},
			shouldErr: true,
		},
		{
			name:       "preflight checks registered, unknown preflight check specified, error returned",
			preflights: "nonexistent",
			registry: &Registry{
				known: map[string]Check{
					"exists": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
				},
				enabledFlag: map[string]bool{},
			},
			shouldErr: true,
		},
		{
			name:       "preflight checks registered, valid input, no error returned",
			preflights: "someCheck",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
				},
				enabledFlag: map[string]bool{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Set(tc.preflights)
			require.Equalf(t, tc.shouldErr, err != nil, "Unexpected error: %v", err)
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
					"disabledCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return errors.New("should be disabled")
					}, nil, false),
				},
			},
		},
		{
			name: "preflight checks registered, enabled check returns an error, error returned",
			registry: &Registry{
				known: map[string]Check{
					"errorCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return errors.New("error")
					}, nil, true),
				},
			},
			shouldErr: true,
		},
		{
			name: "preflight checks registered, enabled checks successful, no error returned",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(func(_ context.Context, _ *diffgraph.ChangeGraph, _ CheckConfig) error {
						return nil
					}, nil, true),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.registry.Run(nil, nil)
			require.Equalf(t, tc.shouldErr, err != nil, "Unexpected error: %v", err)
		})
	}
}

func TestRegistryConfig(t *testing.T) {
	testCases := []struct {
		name       string
		registry   *Registry
		configYaml string
		shouldErr  bool
	}{
		{
			name: "preflight checks registered, config set, no error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(
						nil,
						func(cfg CheckConfig) error {
							if cfg == nil {
								return errors.New("config should be present")
							}
							v, ok := cfg["foo"]
							if !ok {
								return errors.New("foo config not present")
							}
							if v != "bar" {
								return errors.New("foo should equal 'bar'")
							}
							_, ok = cfg["foobar"]
							if ok {
								return errors.New("foobar config should not present")
							}
							return nil
						},
						true,
					),
				},
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
  config:
    foo: bar
`,
		},
		{
			name: "preflight checks registered, unexpected preflight check set, error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(
						nil,
						func(cfg CheckConfig) error {
							if cfg != nil {
								return errors.New("config should not be present")
							}
							return nil
						},
						true,
					),
				},
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: otherCheck
  config:
    foo: bar
`,
			shouldErr: true,
		},
		{
			name: "preflight checks registered, duplicate config entry, error",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(nil, nil, true),
				},
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
  config:
    foo: bar
- name: someCheck
  config:
    bar: foo
`,
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configRs, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(tc.configYaml))).Resources()
			require.NoErrorf(t, err, "Parsing resources")

			_, conf, err := ctlconf.NewConfFromResources(configRs)
			require.NoErrorf(t, err, "Parsing config")

			err = tc.registry.SetConfig(conf.PreflightRules())
			require.Equalf(t, tc.shouldErr, err != nil, "Unexpected error: %v", err)
		})
	}
}

func TestRegistryEnable(t *testing.T) {
	testCases := []struct {
		name       string
		registry   *Registry
		results    map[string]bool
		cmdEnable  string
		configYaml string
	}{
		{
			name: "preflight enabled via config and command line",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(nil, nil, false),
				},
				enabledFlag: map[string]bool{},
			},
			results: map[string]bool{
				"someCheck": true,
			},
			cmdEnable: "someCheck",
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
`,
		},
		{
			name: "preflight enabled via config, no command line",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(nil, nil, false),
				},
				enabledFlag: map[string]bool{},
			},
			results: map[string]bool{
				"someCheck": true,
			},
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
`,
		},
		{
			name: "preflight enabled via command line, no config",
			registry: &Registry{
				known: map[string]Check{
					"someCheck": NewCheck(nil, nil, false),
				},
				enabledFlag: map[string]bool{},
			},
			results: map[string]bool{
				"someCheck": true,
			},
			cmdEnable: "someCheck",
		},
		{
			name: "preflight enabled via config, disabled via command line",
			registry: &Registry{
				known: map[string]Check{
					"someCheck":  NewCheck(nil, nil, false),
					"otherCheck": NewCheck(nil, nil, false),
				},
				enabledFlag: map[string]bool{},
			},
			results: map[string]bool{
				"someCheck":  false,
				"otherCheck": true,
			},
			cmdEnable: "otherCheck", // someCheck not listed, so it's disabled
			configYaml: `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
preflightRules:
- name: someCheck
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Enable from commandline first
			if tc.cmdEnable != "" {
				err := tc.registry.Set(tc.cmdEnable)
				require.NoErrorf(t, err, "Unexpected error from Set(): %v", err)
			}
			// Enable from config file second
			if tc.configYaml != "" {
				configRs, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(tc.configYaml))).Resources()
				require.NoErrorf(t, err, "Parsing resources")
				_, conf, err := ctlconf.NewConfFromResources(configRs)
				require.NoErrorf(t, err, "Parsing config")
				err = tc.registry.SetConfig(conf.PreflightRules())
				require.NoErrorf(t, err, "Unexpected error from SetConfig(): %v", err)
			}
			// Check enable results
			for k, v := range tc.results {
				require.Equalf(t, v, tc.registry.known[k].Enabled(), "Unexpected enable value")
			}
		})
	}
}
