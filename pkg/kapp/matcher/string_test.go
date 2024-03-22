// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package matcher_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/matcher"
)

func TestStringMatcherMatches(t *testing.T) {
	exs := []stringMatcherExample{
		{Expected: "app", Actual: "app", Result: true},
		{Expected: "app", Actual: "ap", Result: false},

		{Expected: "app*", Actual: "app", Result: true},
		{Expected: "app*", Actual: "app-extra", Result: true},
		{Expected: "app*", Actual: "ap", Result: false},
		{Expected: "app*", Actual: "ap-extra", Result: false},

		{Expected: "*app", Actual: "app", Result: true},
		{Expected: "*app", Actual: "extra-app", Result: true},
		{Expected: "*app", Actual: "pp", Result: false},
		{Expected: "*app", Actual: "extra-pp", Result: false},

		{Expected: "*app*", Actual: "app", Result: true},
		{Expected: "*app*", Actual: "ap", Result: false},
		{Expected: "*app*", Actual: "app-extra", Result: true},
		{Expected: "*app*", Actual: "ap-extra", Result: false},
		{Expected: "*app*", Actual: "extra-app", Result: true},
		{Expected: "*app*", Actual: "extra-pp", Result: false},
		{Expected: "*app*", Actual: "extra-app-extra", Result: true},
		{Expected: "*app*", Actual: "extra-ap-extra", Result: false},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type stringMatcherExample struct {
	Expected string
	Actual   string
	Result   bool
}

func (e stringMatcherExample) Check(t *testing.T) {
	result := matcher.NewStringMatcher(e.Expected).Matches(e.Actual)
	require.Equal(t, e.Result, result, "Did not match result: expected=%s actual=%s", e.Expected, e.Actual)
}
