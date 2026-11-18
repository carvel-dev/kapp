// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"testing"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestCompactBytesLength(t *testing.T) {
	fullBs := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  namespace: ns
  annotations:
    # plenty of indented values to add whitespace chars
    ann1: ann-val
    ann2: ann-val
    ann3: ann-val
    ann4: ann-val
    ann5: ann-val
    ann6: ann-val
    ann7: ann-val
    ann8: ann-val
    ann9: ann-val
    ann10: ann-val
    ann11: ann-val
    ann12: ann-val
    ann13: |
      val1
      val2
      val3
`

	// resource can be read from full repr
	resFromFull, err := ctlres.NewResourceFromBytes([]byte(fullBs))
	require.NoError(t, err, "Expected to parse full bytes")

	compactBs, err := resFromFull.AsCompactBytes()
	require.NoError(t, err, "Expected to produce compact bytes")

	// resource can be read from compact repr
	resFromCompact, err := ctlres.NewResourceFromBytes([]byte(compactBs))
	require.NoError(t, err, "Expected to parse compact bytes")

	require.True(t, resFromFull.Equal(resFromCompact), "Expected resources to match: %q vs %q", fullBs, compactBs)

	require.Less(t, len(compactBs), len(fullBs), "Compact repr should be shorter than full repr")
}

func TestCompactBytesNoNewlinesForBetterFormatting(t *testing.T) {
	fullBs := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  annotations:
    ann13: |
      val1
      val2
      val3
`

	resFromFull, err := ctlres.NewResourceFromBytes([]byte(fullBs))
	require.NoError(t, err, "Expected to parse full bytes")

	compactBs, err := resFromFull.AsCompactBytes()
	require.NoError(t, err, "Expected to produce compact bytes")

	require.Contains(t, string(fullBs), "\n", "Expected full repr to have newlines")

	require.NotContains(t, string(compactBs), "\n", "Expected compact repr to not have newlines")
}

func TestGroupVersion(t *testing.T) {
	tests := []struct {
		name            string
		apiVersion      string
		expectedGroup   string
		expectedVersion string
	}{
		{"core group", "v1", "", "v1"},
		{"named group", "apps/v1", "apps", "v1"},
		{"more separators than a group and a version", "a/b/c", "", ""},
		{"empty", "", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: ` + test.apiVersion + `
kind: Config
metadata:
  name: cfg
`))
			gv := res.GroupVersion()
			require.Equal(t, test.expectedGroup, gv.Group)
			require.Equal(t, test.expectedVersion, gv.Version)
		})
	}
}
