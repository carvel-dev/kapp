/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources_test

import (
	"strings"
	"testing"
)

func expectEqualsStripped(t *testing.T, description, resultStr, expectedStr string) {
	expectEquals(t, description, strings.TrimSpace(resultStr), strings.TrimSpace(expectedStr))
}

func expectEquals(t *testing.T, description, resultStr, expectedStr string) {
	if resultStr != expectedStr {
		t.Fatalf("%s: not equal\n\n### result %d chars:\n>>>%s<<<\n###expected %d chars:\n>>>%s<<<",
			description, len(resultStr), resultStr, len(expectedStr), expectedStr)
	}
}
