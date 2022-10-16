// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

func TestStripNameHashSuffix_TestConfigIncludeExclude(t *testing.T) {

	includedCM := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-abc
data:
  foo: foo
`))

	excludedCM := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: foo
data:
  foo: foo
`))

	includedKind := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v2
kind: MyKind
metadata:
  name: my-res
spec:
  key: val
`))

	excludedKind := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
spec:
  replicas: 1
`))

	includes := []ctlconf.ResourceMatcher{
		ctlconf.ResourceMatcher{
			APIVersionKindMatcher: &ctlconf.APIVersionKindMatcher{APIVersion: includedCM.APIVersion(), Kind: includedCM.Kind()},
		},
		ctlconf.ResourceMatcher{
			APIVersionKindMatcher: &ctlconf.APIVersionKindMatcher{APIVersion: includedKind.APIVersion(), Kind: includedKind.Kind()},
		},
	}

	excludes := []ctlconf.ResourceMatcher{
		ctlconf.ResourceMatcher{
			KindNamespaceNameMatcher: &ctlconf.KindNamespaceNameMatcher{Kind: excludedCM.Kind(), Name: excludedCM.Name()},
		},
	}

	configCases := []struct {
		desc  string
		confs []ctlconf.StripNameHashSuffixConfig
	}{
		{
			"SingleConfig",
			[]ctlconf.StripNameHashSuffixConfig{
				ctlconf.StripNameHashSuffixConfig{
					Includes: includes,
					Excludes: excludes,
				},
			},
		},
		{
			"AccrossConfigs",
			[]ctlconf.StripNameHashSuffixConfig{
				ctlconf.StripNameHashSuffixConfig{
					Includes: includes,
				},
				ctlconf.StripNameHashSuffixConfig{
					Excludes: excludes,
				},
			},
		},
	}

	for _, configTc := range configCases {
		t.Run(configTc.desc, func(t *testing.T) {

			config := newStripNameHashSuffixConfigFromConf(configTc.confs)

			resCases := []struct {
				kind     string
				res      ctlres.Resource
				expected bool
			}{
				{"ConfigMap", includedCM, true},
				{"ConfigMap", excludedCM, false},
				{"Kind", includedKind, true},
				{"Kind", excludedKind, false},
			}

			for _, resTc := range resCases {
				var inEx, not string
				if resTc.expected {
					inEx = "included"
					not = ""
				} else {
					inEx = "excluded"
					not = " not"
				}
				desc := fmt.Sprintf("%s_%s", resTc.kind, inEx)
				t.Run(desc, func(t *testing.T) {
					require.Equalf(t, resTc.expected, config.EnabledFor(resTc.res), "expected %s %s to%s match!", inEx, resTc.kind, not)
				})
			}
		})
	}
}

func TestStripNameHashSuffix_TestConfig_DefaultNone(t *testing.T) {
	excludedCM := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: foo
data:
  foo: foo
`))

	excludedKind := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
spec:
  replicas: 1
`))

	config := newStripNameHashSuffixConfigFromConf([]ctlconf.StripNameHashSuffixConfig{})

	cases := []struct {
		kind string
		res  ctlres.Resource
	}{
		{"ConfigMap", excludedCM},
		{"Deployment", excludedKind},
	}

	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			require.Falsef(t, config.EnabledFor(excludedCM), "expected %s to not match by default!")
		})
	}
}
