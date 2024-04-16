// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"testing"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/stretchr/testify/require"
)

// validateChanges: common func used by multiple test cases for validation between actual and expected
func validateChanges(t *testing.T, respTable []ui.JSONUITableResp, expected []map[string]string, notesOp string,
	notesWaitTo string, output string) {

	//deleting age from response table rows as it is varying from 0s to 1s making test case fail
	for _, row := range respTable[0].Rows {
		delete(row, "age")
	}

	require.Exactlyf(t, expected, respTable[0].Rows, "Expected to see correct changes, but did not: '%s'", output)
	require.Equalf(t, notesOp, respTable[0].Notes[0], "Expected to see correct summary, but did not: '%s'", output)
	require.Equalf(t, notesWaitTo, respTable[0].Notes[1], "Expected to see correct summary, but did not: '%s'", output)

}

func replaceAge(result []map[string]string) []map[string]string {
	for i, row := range result {
		if len(row["age"]) > 0 {
			row["age"] = "<replaced>"
		}
		result[i] = row
	}
	return result
}

func replaceLastChangeAge(result []map[string]string) []map[string]string {
	for i, row := range result {
		if len(row["last_change_age"]) > 0 {
			row["last_change_age"] = "<replaced>"
		}
		result[i] = row
	}
	return result
}

func replaceAnnsLabels(in string) string {
	replaceAnns := regexp.MustCompile("kapp\\.k14s\\.io\\/(app|association): .+")
	return replaceAnns.ReplaceAllString(in, "-replaced-")
}
