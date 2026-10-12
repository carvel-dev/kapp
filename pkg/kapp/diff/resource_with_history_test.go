// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDebugEnabled(t *testing.T) {
	exs := []debugEnabledExample{
		{Val: "true", Result: true},
		{Val: "TRUE", Result: true},
		{Val: "True", Result: true},
		{Val: "tRuE", Result: true},

		{Val: "", Result: false},
		{Val: "false", Result: false},
		{Val: "1", Result: false},
		{Val: "yes", Result: false},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type debugEnabledExample struct {
	Val    string
	Result bool
}

func (e debugEnabledExample) Check(t *testing.T) {
	require.Equal(t, e.Result, debugEnabled(e.Val), "Did not match result: val=%q", e.Val)
}
