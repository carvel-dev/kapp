// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"reflect"
	"testing"

	"github.com/cppforlife/go-cli-ui/ui"
)

// validateChanges: common func used by multiple test cases for validation between actual and expected
func validateChanges(t *testing.T, respTable []ui.JSONUITableResp, expected []map[string]string, notesOp string,
	notesWaitTo string, output string) {

	//deleting age from response table rows as it is varying from 0s to 1s making test case fail
	for _, row := range respTable[0].Rows {
		delete(row, "age")
	}

	if !reflect.DeepEqual(respTable[0].Rows, expected) {
		t.Fatalf("Expected to see correct changes, but did not: '%s'", output)
	}
	if respTable[0].Notes[0] != notesOp {
		t.Fatalf("Expected to see correct summary, but did not: '%s'", output)
	}
	if respTable[0].Notes[1] != notesWaitTo {
		t.Fatalf("Expected to see correct summary, but did not: '%s'", output)
	}
}
