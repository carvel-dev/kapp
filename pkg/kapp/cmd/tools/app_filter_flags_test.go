// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"testing"
)

func TestAppFilterFlagsTimes(t *testing.T) {
	exs := []filterFlagsTimesExample{
		{Age: "", Before: false, After: false},
		{Age: "5m+", Before: true, After: false},
		{Age: "5m-", Before: false, After: true},
		{Age: "not-a-duration", ExpectedErr: invalidAgeFilterErr},
	}

	for _, ex := range exs {
		ex.Check(t, func() (bool, bool, error) {
			before, after, err := (&AppFilterFlags{age: ex.Age}).Times()
			return before != nil, after != nil, err
		})
	}
}
