// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestCRDVersionsWithGKScoping(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	crdYamlTemplate := `
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: foostores.demo.com
spec:
  group: demo.com
  versions:
    - name: v1beta1
      served: %s
      storage: false
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                foo:
                  type: string
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                foo:
                  type: string
  scope: Namespaced
  names:
    plural: foostores
    singular: foostore
    kind: FooStore
    shortNames:
    - fst	
`

	crdWithV1beta1Served := fmt.Sprintf(crdYamlTemplate, "true")
	crdWithV1beta1NotServed := fmt.Sprintf(crdYamlTemplate, "false")

	crYaml := `
---
apiVersion: demo.com/v1beta1
kind: FooStore
metadata:
  name: test-cr
spec:
  foo: bar1
`
	name := "test-crd-versions-with-gk-scoping"

	// CRD with versions v1beta1 and v1 served
	// CR using version v1beta1
	yaml1 := fmt.Sprintf("%s\n%s", crdWithV1beta1Served, crYaml)

	// Updating CRD to stop serving version v1beta1
	yaml2 := fmt.Sprintf("%s\n%s", crdWithV1beta1NotServed, crYaml)

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}
	defer cleanUp()

	logger.Section("initial deploy", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-", "--json"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{
			{
				"age":             "",
				"kind":            "CustomResourceDefinition",
				"name":            "foostores.demo.com",
				"namespace":       "(cluster)",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
			{
				"age":             "",
				"kind":            "FooStore",
				"name":            "test-cr",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
		}

		require.Exactlyf(t, expected, respRows, "Expected to see correct changes")
	})

	logger.Section("stop serving CRD on v1beta1", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-", "--json"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{
			{
				"age":             "<replaced>",
				"kind":            "CustomResourceDefinition",
				"name":            "foostores.demo.com",
				"namespace":       "(cluster)",
				"op":              "update",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "ok",
				"wait_to":         "reconcile",
			},
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})

	logger.Section("remove cr from manifest", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-", "--json"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(crdWithV1beta1NotServed)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		respRows := resp.Tables[0].Rows

		expected := []map[string]string{
			{
				"age":             "<replaced>",
				"kind":            "FooStore",
				"name":            "test-cr",
				"namespace":       "kapp-test",
				"op":              "delete",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "ok",
				"wait_to":         "delete",
			},
		}

		require.Exactlyf(t, expected, replaceAge(respRows), "Expected to see correct changes")
	})

}
