// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

// test that we can exclude specific configsmaps from default
func TestStripNameHashSuffix_TestConfig_IncludeExcludeAccrossConfigs(t *testing.T) {
	requireStripNameHashSuffixMatches(t, [][]ctlres.ResourceMatcher{
		[]ctlres.ResourceMatcher{
			ctlres.AnyMatcher {
				Matchers: []ctlres.ResourceMatcher {
					ctlres.APIVersionKindMatcher{APIVersion: "v1", Kind: "ConfigMap"},
					ctlres.APIVersionKindMatcher{APIVersion: "v2", Kind: "MyKind"},
				},
			},
		},
		[]ctlres.ResourceMatcher{
			ctlres.NotMatcher{
				Matcher: ctlres.KindNamespaceNameMatcher{Kind: "ConfigMap", Name: "foo"},
			},
		},
	})
}

func TestStripNameHashSuffix_TestConfig_IncludeExcludeSingleConfig(t *testing.T) {
	requireStripNameHashSuffixMatches(t, [][]ctlres.ResourceMatcher{
		[]ctlres.ResourceMatcher{
			ctlres.AnyMatcher {
				Matchers: []ctlres.ResourceMatcher {
					ctlres.APIVersionKindMatcher{APIVersion: "v1", Kind: "ConfigMap"},
					ctlres.APIVersionKindMatcher{APIVersion: "v2", Kind: "MyKind"},
				},
			},
			ctlres.NotMatcher{
				Matcher: ctlres.KindNamespaceNameMatcher{Kind: "ConfigMap", Name: "foo"},
			},
		},
	})
}

func requireStripNameHashSuffixMatches(t *testing.T, matchers [][]ctlres.ResourceMatcher) {
	includedCM := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-abc
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

	config := newStripNameHashSuffixConfig(true, matchers)

	require.True(t, config.EnabledFor(includedCM), "expected included ConfigMap to match!")
	require.True(t, config.EnabledFor(includedKind), "expected included Kind to match!")
	require.False(t, config.EnabledFor(excludedCM), "expected excluded ConfigMap to not match!")
	require.False(t, config.EnabledFor(excludedKind), "expected unmentioned Kind to not match!")

}
