// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
