package e2e

import (
	"reflect"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestIgnoreFailingAPIServices(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1.dummykapptest.com
  annotations:
    kapp.k14s.io/disable-default-change-group-and-rules: ""
    kapp.k14s.io/change-group: "apiservice"
spec:
  group: dummykapptest.com
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: redis-master
    namespace: kapp-test
  version: v1
  versionPriority: 100
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: foo.dummykapptest.com
  annotations:
    kapp.k14s.io/disable-default-change-group-and-rules: ""
    kapp.k14s.io/change-rule: "upsert after upserting apiservice"
spec:
  group: dummykapptest.com
  versions:
  - name: v1
    served: true
    storage: true
  scope: Namespaced
  names:
    plural: foo
    singular: foo
    kind: Foo
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-ignore-failing-api-service
`

	yaml3 := `
---
apiVersion: dummykapptest.com/v1
kind: Foo
metadata:
  name: test-uses-failing-api-service
`

	name1 := "test-ignore-failing-api-services1"
	name2 := "test-ignore-failing-api-services2"
	name3 := "test-ignore-failing-api-services3"

	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name1}, RunOpts{AllowError: true})
		kapp.RunWithOpts([]string{"delete", "-a", name2}, RunOpts{AllowError: true})
		kapp.RunWithOpts([]string{"delete", "-a", name3}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy broken api service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name1, "--wait=false"}, RunOpts{
			IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy app that does not use api service", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2}, RunOpts{
			IntoNs: true, StdinReader: strings.NewReader(yaml2)})
	})

	logger.Section("inspect app that does not use api service", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name2, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "test-ignore-failing-api-service",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		if !reflect.DeepEqual(replaceAge(resp.Tables[0].Rows), expected) {
			t.Fatalf("Expected to see correct changes, but did not: '%s'", out)
		}
	})

	logger.Section("deploy app that uses failing api service", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name3}, RunOpts{
			AllowError: true, IntoNs: true, StdinReader: strings.NewReader(yaml3)})
		if err == nil {
			t.Fatalf("Expected error when deploying with failing api service")
		}
		if !strings.Contains(err.Error(), "unable to retrieve the complete list of server APIs: dummykapptest.com/v1: the server is currently unable to handle the request") {
			t.Fatalf("Expected api retrieval error but was '%s'", err)
		}
	})

	logger.Section("deploy app that uses failing api service and try to ignore it", func() {
		ignoreFlag := "--dangerous-ignore-failing-api-services"

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name3, ignoreFlag}, RunOpts{
			AllowError: true, IntoNs: true, StdinReader: strings.NewReader(yaml3)})
		if err == nil {
			t.Fatalf("Expected error when deploying with failing api service")
		}
		if !strings.Contains(err.Error(), "Expected to find kind 'dummykapptest.com/v1/Foo', but did not") {
			t.Fatalf("Expected CRD retrieval error but was '%s'", err)
		}
	})

	logger.Section("delete app that does not user api service", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name2}, RunOpts{})
	})

	logger.Section("delete failing api service", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name1}, RunOpts{})
	})
}
