// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
	if err != nil {
		t.Fatalf("Expected to parse full bytes: %s", err)
	}

	compactBs, err := resFromFull.AsCompactBytes()
	if err != nil {
		t.Fatalf("Expected to produce compact bytes: %s", err)
	}

	// resource can be read from compact repr
	resFromCompact, err := ctlres.NewResourceFromBytes([]byte(compactBs))
	if err != nil {
		t.Fatalf("Expected to parse compact bytes: %s", err)
	}

	if !resFromFull.Equal(resFromCompact) {
		t.Fatalf("Expected resources to match: '%s' vs '%s'", fullBs, compactBs)
	}

	if len(compactBs) >= len(fullBs) {
		t.Fatalf("Compact repr should be shorter than full repr: %d '%s' vs %d '%s'",
			len(fullBs), fullBs, len(compactBs), compactBs)
	}
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
	if err != nil {
		t.Fatalf("Expected to parse full bytes: %s", err)
	}

	compactBs, err := resFromFull.AsCompactBytes()
	if err != nil {
		t.Fatalf("Expected to produce compact bytes: %s", err)
	}

	if !strings.Contains(string(fullBs), "\n") {
		t.Fatalf("Expected full repr to have newlines: %s", fullBs)
	}
	if strings.Contains(string(compactBs), "\n") {
		t.Fatalf("Expected compact repr to not have newlines: %s", compactBs)
	}
}
