// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
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
    name: redis-primary
    namespace: kapp-test
  version: v1
  versionPriority: 100
---
apiVersion: apiextensions.k8s.io/v1
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
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
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
		kapp.Run([]string{"delete", "-a", name1})
		kapp.Run([]string{"delete", "-a", name2})
		kapp.Run([]string{"delete", "-a", name3})
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

		require.Exactlyf(t, expected, replaceAge(resp.Tables[0].Rows), "Expected to see correct changes")
	})

	logger.Section("deploy app that uses failing api service", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name3}, RunOpts{
			AllowError: true, IntoNs: true, StdinReader: strings.NewReader(yaml3)})
		require.Errorf(t, err, "Expected error when deploying with failing api service")

		require.Contains(t, err.Error(), "unable to retrieve the complete list of server APIs: dummykapptest.com/v1: the server is currently unable to handle the request",
			"Expected api retrieval error")
	})

	logger.Section("deploy app that uses failing api service and try to ignore it", func() {
		ignoreFlag := "--dangerous-ignore-failing-api-services"

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name3, ignoreFlag}, RunOpts{
			AllowError: true, IntoNs: true, StdinReader: strings.NewReader(yaml3)})
		require.Errorf(t, err, "Expected error when deploying with failing api service")

		require.Contains(t, err.Error(), "Expected to find kind 'dummykapptest.com/v1/Foo', but did not", "Expected CRD retrieval error")
	})

	logger.Section("delete app that does not use api service", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name2}, RunOpts{})
	})

	logger.Section("delete failing api service", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name1}, RunOpts{})
	})
}

func TestIgnoreFailingGroupVersion(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: foo.dummykapptest.com
spec:
  group: dummykapptest.com
  versions:
  # v1 is available and used for internal storage
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
  # v2 is available but gets converted from v1 on the fly
  - name: v2
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
  scope: Namespaced
  names:
    plural: foo
    singular: foo
    kind: Foo
  preserveUnknownFields: false
  conversion:
    strategy: Webhook
    webhook:
      conversionReviewVersions: ["v1","v1beta1"]
      clientConfig:
        service:
          namespace: kapp-test
          name: failing-group-version-webhook
          path: /convert
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-ignore-failing-group-version
`

	yaml3 := `
---
apiVersion: dummykapptest.com/v2
kind: Foo
metadata:
  name: test-uses-failing-group-version
spec: {}
`

	name1 := "test-ignore-failing-group-version1"
	name2 := "test-ignore-failing-group-version2"
	name3 := "test-ignore-failing-group-version3"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name1})
		kapp.Run([]string{"delete", "-a", name2})
		kapp.Run([]string{"delete", "-a", name3})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy broken CRD", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name1, "--wait=false"}, RunOpts{
			IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy app that does not use failing group version", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2}, RunOpts{
			IntoNs: true, StdinReader: strings.NewReader(yaml2)})
	})

	logger.Section("inspect app that does not use failing group version", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name2, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "test-ignore-failing-group-version",
			"namespace":       "kapp-test",
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		require.Exactlyf(t, expected, replaceAge(resp.Tables[0].Rows), "Expected to see correct changes")
	})

	logger.Section("deploy app that uses failing group version", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name3, "--apply-timeout=5s"}, RunOpts{
			AllowError: true, IntoNs: true, StdinReader: strings.NewReader(yaml3)})
		require.Errorf(t, err, "Expected error when deploying with failing group version")

		require.Contains(t, err.Error(), `service "failing-group-version-webhook" not found`, "Expected api retrieval error")
	})

	logger.Section("delete app that uses failing group version", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name3}, RunOpts{})
	})

	logger.Section("delete app that does not use failing group version", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name2}, RunOpts{})
	})

	logger.Section("delete app that does not use failing group version", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name2}, RunOpts{})
	})
}
