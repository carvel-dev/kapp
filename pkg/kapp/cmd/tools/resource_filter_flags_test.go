// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceFilterFlagsTimes(t *testing.T) {
	exs := []filterFlagsTimesExample{
		{Age: "", Before: false, After: false},
		{Age: "5m+", Before: true, After: false},
		{Age: "5m-", Before: false, After: true},
		{Age: "not-a-duration", ExpectedErr: invalidAgeFilterErr},
	}

	for _, ex := range exs {
		ex.Check(t, func() (bool, bool, error) {
			before, after, err := (&ResourceFilterFlags{Age: ex.Age}).Times()
			return before != nil, after != nil, err
		})
	}
}

const invalidAgeFilterErr = "Expected age filter to be either empty or " +
	"a valid time.Duration (example: 5m+, 24h-; valid units: ns, us, ms, s, m, h)"

type filterFlagsTimesExample struct {
	Age         string
	Before      bool
	After       bool
	ExpectedErr string
}

func (e filterFlagsTimesExample) Check(t *testing.T, times func() (bool, bool, error)) {
	before, after, err := times()
	if len(e.ExpectedErr) > 0 {
		require.EqualError(t, err, e.ExpectedErr, "age=%s", e.Age)
		return
	}

	require.NoError(t, err, "age=%s", e.Age)
	require.Equal(t, e.Before, before, "Unexpected before time: age=%s", e.Age)
	require.Equal(t, e.After, after, "Unexpected after time: age=%s", e.Age)
}
