// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestListOptionsParseTime(t *testing.T) {
	exs := []parseTimeExample{
		{Input: "2024-03-05", Expected: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)},
		{Input: "2024-03-05T10:20:30Z", Expected: time.Date(2024, 3, 5, 10, 20, 30, 0, time.UTC)},
		{Input: "yesterday", ExpectedErr: "Unrecognized time format yesterday"},
		{Input: "03/05/2024", ExpectedErr: "Unrecognized time format 03/05/2024"},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type parseTimeExample struct {
	Input       string
	Expected    time.Time
	ExpectedErr string
}

func (e parseTimeExample) Check(t *testing.T) {
	formats := []string{time.RFC3339, "2006-01-02"}

	result, err := (&ListOptions{}).parseTime(e.Input, formats)
	if len(e.ExpectedErr) > 0 {
		require.Error(t, err, "input=%s", e.Input)
		require.Contains(t, err.Error(), e.ExpectedErr, "input=%s", e.Input)
		return
	}

	require.NoError(t, err, "input=%s", e.Input)
	require.Equal(t, e.Expected, result, "input=%s", e.Input)
}
