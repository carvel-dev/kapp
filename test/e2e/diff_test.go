// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestDiff(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config1
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config1
data:
  key: value2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config3
data:
  key: value
`

	name := "test-diff"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "",
			"op":              "create",
			"op_strategy":     "",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "",
		}, {
			"age":             "",
			"op":              "create",
			"op_strategy":     "",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config1",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "",
		}, {
			"age":             "",
			"op":              "create",
			"op_strategy":     "",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config2",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "",
		}}

		if !reflect.DeepEqual(resp.Tables[0].Rows, expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "Op:      3 create, 0 delete, 0 update, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[1] != "Wait to: 3 reconcile, 0 delete, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})

	logger.Section("deploy no change", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		expected := []map[string]string{}

		if !reflect.DeepEqual(resp.Tables[0].Rows, expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "Op:      0 create, 0 delete, 0 update, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[1] != "Wait to: 0 reconcile, 0 delete, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})

	logger.Section("deploy update with 1 delete, 1 update, 1 create", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "update",
			"op_strategy":     "",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config1",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "",
			"op":              "create",
			"op_strategy":     "",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config3",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "Op:      1 create, 1 delete, 1 update, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[1] != "Wait to: 2 reconcile, 1 delete, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})

	logger.Section("delete", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config1",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config2",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "redis-config3",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "Op:      0 create, 3 delete, 0 update, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[1] != "Wait to: 0 reconcile, 3 delete, 0 noop" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})
}

func TestDiffExitStatus(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	name := "test-diff-exit-status"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
		"--diff-run", "--diff-exit-status", "--dangerous-allow-empty-list-of-resources"},
		RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader("---\n")})
	if err == nil {
		t.Fatalf("Expected to receive error")
	}

	if !strings.Contains(err.Error(), "Exiting after diffing with no pending changes (exit status 2)") {
		t.Fatalf("Expected to find stderr output")
	}
	if !strings.Contains(err.Error(), "exit code: '2'") {
		t.Fatalf("Expected to find exit code")
	}

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
`

	_, err = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
		"--diff-run", "--diff-exit-status"},
		RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})
	if err == nil {
		t.Fatalf("Expected to receive error")
	}

	if !strings.Contains(err.Error(), "Exiting after diffing with pending changes (exit status 3)") {
		t.Fatalf("Expected to find stderr output")
	}
	if !strings.Contains(err.Error(), "exit code: '3'") {
		t.Fatalf("Expected to find exit code")
	}
}

func TestDiffMaskRules(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	yaml1 := `
apiVersion: v1
kind: Secret
metadata:
  name: no-data
---
apiVersion: v1
kind: Secret
metadata:
  name: empty-data
data: {}
---
apiVersion: v1
kind: Secret
metadata:
  name: with-keys
data:
  key1: val1
  key2: val2
---
apiVersion: v1
kind: Secret
metadata:
  name: with-dup-keys
data:
  key1: val1
  key2: val2
`

	yaml2 := `
---
apiVersion: v1
kind: Secret
metadata:
  name: with-dup-keys
data:
  key1: val1
  key2: val3
`

	name := "test-diff-mask-rules"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c"},
		RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

	expectedOutput := `
@@ create secret/no-data (v1) namespace: kapp-test @@
      0 + apiVersion: v1
      1 + kind: Secret
      2 + metadata:
      3 +   labels:
      4 +     -replaced-
      5 +     -replaced-
      6 +   name: no-data
      7 +   namespace: kapp-test
      8 + 
@@ create secret/empty-data (v1) namespace: kapp-test @@
      0 + apiVersion: v1
      1 + data: {}
      2 + kind: Secret
      3 + metadata:
      4 +   labels:
      5 +     -replaced-
      6 +     -replaced-
      7 +   name: empty-data
      8 +   namespace: kapp-test
      9 + 
@@ create secret/with-keys (v1) namespace: kapp-test @@
      0 + apiVersion: v1
      1 + data:
      2 +   key1: <-- value not shown (#1)
      3 +   key2: <-- value not shown (#2)
      4 + kind: Secret
      5 + metadata:
      6 +   labels:
      7 +     -replaced-
      8 +     -replaced-
      9 +   name: with-keys
     10 +   namespace: kapp-test
     11 + 
@@ create secret/with-dup-keys (v1) namespace: kapp-test @@
      0 + apiVersion: v1
      1 + data:
      2 +   key1: <-- value not shown (#1)
      3 +   key2: <-- value not shown (#2)
      4 + kind: Secret
      5 + metadata:
      6 +   labels:
      7 +     -replaced-
      8 +     -replaced-
      9 +   name: with-dup-keys
     10 +   namespace: kapp-test
     11 + 
`

	out = replaceAnnsLabels(out)

	if !strings.Contains(out, expectedOutput) {
		t.Fatalf("Did not find expected diff output >>%s<< in >>%s<<", out, expectedOutput)
	}

	out, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "-p"},
		RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

	expectedOutput = `
@@ update secret/with-dup-keys (v1) namespace: kapp-test @@
  ...
  2,  2     key1: <-- value not shown (#1)
  3     -   key2: <-- value not shown (#2)
      3 +   key2: <-- value not shown (#3)
  4,  4   kind: Secret
  5,  5   metadata:
`

	if !strings.Contains(out, expectedOutput) {
		t.Fatalf("Did not find expected diff output >>%s<< in >>%s<<", out, expectedOutput)
	}
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

func replaceAnnsLabels(in string) string {
	replaceAnns := regexp.MustCompile("kapp\\.k14s\\.io\\/(app|association): .+")
	return replaceAnns.ReplaceAllString(in, "-replaced-")
}
