package e2e

import (
	"reflect"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestDiff(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, logger}

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
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":        "",
			"op":         "create",
			"wait_to":    "reconcile",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config",
			"namespace":  "kapp-test",
		}, {
			"age":        "",
			"op":         "create",
			"wait_to":    "reconcile",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config1",
			"namespace":  "kapp-test",
		}, {
			"age":        "",
			"op":         "create",
			"wait_to":    "reconcile",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config2",
			"namespace":  "kapp-test",
		}}

		if !reflect.DeepEqual(resp.Tables[0].Rows, expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "3 create, 0 delete, 0 update" {
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
		if resp.Tables[0].Notes[0] != "0 create, 0 delete, 0 update" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})

	logger.Section("deploy update with 1 delete, 1 update, 1 create", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":        "<replaced>",
			"op":         "delete",
			"wait_to":    "delete",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config",
			"namespace":  "kapp-test",
		}, {
			"age":        "<replaced>",
			"op":         "update",
			"wait_to":    "reconcile",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config1",
			"namespace":  "kapp-test",
		}, {
			"age":        "",
			"op":         "create",
			"wait_to":    "reconcile",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config3",
			"namespace":  "kapp-test",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "1 create, 1 delete, 1 update" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})

	logger.Section("delete", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":        "<replaced>",
			"op":         "delete",
			"wait_to":    "delete",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config1",
			"namespace":  "kapp-test",
		}, {
			"age":        "<replaced>",
			"op":         "delete",
			"wait_to":    "delete",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config2",
			"namespace":  "kapp-test",
		}, {
			"age":        "<replaced>",
			"op":         "delete",
			"wait_to":    "delete",
			"conditions": "",
			"kind":       "ConfigMap",
			"name":       "redis-config3",
			"namespace":  "kapp-test",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
		if resp.Tables[0].Notes[0] != "0 create, 3 delete, 0 update" {
			t.Fatalf("Expected to see correct summary, but did not: '%s'", out)
		}
	})
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
