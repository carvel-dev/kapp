package e2e

import (
	"reflect"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestInspect(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-master
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
    role: master
`

	name := "test-inspect"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("plain inspect", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            "redis-master",
			"namespace":       "kapp-test",
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-master",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})

	logger.Section("tree inspect", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "-t", "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Service",
			"name":            "redis-master",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "Endpoints",
			"name":            " L redis-master",
			"namespace":       "kapp-test",
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})
}
