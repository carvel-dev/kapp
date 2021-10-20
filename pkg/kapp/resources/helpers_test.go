// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func expectEqualsStripped(t *testing.T, description, resultStr, expectedStr string) {
	expectEquals(t, description, strings.TrimSpace(resultStr), strings.TrimSpace(expectedStr))
}

func expectEquals(t *testing.T, description, resultStr, expectedStr string) {
	require.Equal(t, expectedStr, resultStr, "%s: not equal", description)
}
