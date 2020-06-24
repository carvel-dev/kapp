package e2e

import (
	"reflect"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestTransientResourceInspectDelete(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-svc
spec:
  selector:
    app: redis-svc
  ports:
  - name: http
    port: 80
`

	name := "test-transient-resource-inspect-delete"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("inspect shows transient resource", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "--json"}, RunOpts{})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}
		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})

	logger.Section("delete includes transient resource", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "delete",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}
		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})
}

func TestTransientResourceSwitchToNonTransient(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-svc
spec:
  selector:
    app: redis-svc
  ports:
  - name: http
    port: 80
`

	yaml2 := yaml1 + `
---
apiVersion: v1
kind: Endpoints
metadata:
  name: redis-svc
`

	name := "test-transient-resource-switch"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy to change transient to non-transient", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "update",
			"wait_to":         "reconcile",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}
		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})

	logger.Section("delete with previously transient resource (now non-transient)", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})
		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "delete",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"op":              "delete",
			"wait_to":         "delete",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-svc",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}
		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})
}
