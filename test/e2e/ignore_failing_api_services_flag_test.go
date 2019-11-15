package e2e

import (
	"reflect"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestIgnoreFailingAPIServicesFlag(t *testing.T) {
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
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1.dummykapptest
spec:
  group: dummykapptest
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: redis-master
    namespace: kapp-test
  version: v1
  versionPriority: 100
`

	ignoreFlag := "--dangerous-ignore-failing-api-services"
	name := "test-ignore-failing-api-services"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name, ignoreFlag}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, ignoreFlag}, RunOpts{
			IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy without flag to see it fail", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{
			AllowError: true, IntoNs: true, StdinReader: strings.NewReader(strings.Replace(yaml1, "6380", "6381", -1))})
		if err == nil {
			t.Fatalf("Expected error when deploying with failing api service")
		}
		if !strings.Contains(err.Error(), "unable to retrieve the complete list of server APIs: dummykapptest/v1: the server is currently unable to handle the request") {
			t.Fatalf("Expected api retrieval error but was '%s'", err)
		}
	})

	logger.Section("deploy with flag to see it succeed", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, ignoreFlag}, RunOpts{
			IntoNs: true, StdinReader: strings.NewReader(strings.Replace(yaml1, "6380", "6381", -1))})
	})

	logger.Section("inspect with flag", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "--json", ignoreFlag}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "0/1 t",
			"kind":            "APIService",
			"name":            "v1.dummykapptest",
			"namespace":       "(cluster)",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
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

	logger.Section("delete with flag", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name, ignoreFlag}, RunOpts{})
	})
}
