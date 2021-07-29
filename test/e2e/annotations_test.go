// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cppforlife/go-cli-ui/ui"
	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	kcv1alpha1 "github.com/k14s/kapp-controller/pkg/apis/kappctrl/v1alpha1"
)

func TestVersionedAnnotations(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  key1: val1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret
data:
  key1: val1
`

	yaml2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
`

	yaml3 := `
--- 
apiVersion: v1
data: 
  key1: val1
kind: ConfigMap
metadata: 
  annotations: 
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/versioned-keep-original: ""
  name: config
--- 
apiVersion: v1
data: 
  key1: val1
kind: Secret
metadata: 
  annotations: 
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/versioned-keep-original: ""
  name: secret
`

	name := "test-versioned-annotations"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Scenario-1 [Non-versioned->Versioned->Versioned-keep-original]", func() {
		nonVerOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		expectedNonVer := []map[string]string{{
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		respNonVer := uitest.JSONUIFromBytes(t, []byte(nonVerOut))

		validateChanges(t, respNonVer.Tables, expectedNonVer, "Op:      2 create, 0 delete, 0 update, 0 noop",
			"Wait to: 2 reconcile, 0 delete, 0 noop", nonVerOut)

		verOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		respVer := uitest.JSONUIFromBytes(t, []byte(verOut))

		expectedVer := []map[string]string{{
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config-ver-1",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret-ver-1",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, respVer.Tables, expectedVer, "Op:      2 create, 2 delete, 0 update, 0 noop",
			"Wait to: 2 reconcile, 2 delete, 0 noop", verOut)

		verKeepOrgOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		respVerKeepOrg := uitest.JSONUIFromBytes(t, []byte(verKeepOrgOut))

		expectedVerKeepOrg := []map[string]string{
			{
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-ver-2",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "Secret",
				"name":            "secret",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "Secret",
				"name":            "secret-ver-2",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
		}

		validateChanges(t, respVerKeepOrg.Tables, expectedVerKeepOrg, "Op:      4 create, 0 delete, 0 update, 0 noop",
			"Wait to: 4 reconcile, 0 delete, 0 noop", verKeepOrgOut)
	})

	cleanUp()

	logger.Section("Scenario-2 [Versioned-keep-original->Versioned]", func() {
		verKeepOrgOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		respVerKeepOrg := uitest.JSONUIFromBytes(t, []byte(verKeepOrgOut))

		expectedVerKeepOrg := []map[string]string{
			{
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "Secret",
				"name":            "secret",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "Secret",
				"name":            "secret-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
		}

		validateChanges(t, respVerKeepOrg.Tables, expectedVerKeepOrg, "Op:      4 create, 0 delete, 0 update, 0 noop",
			"Wait to: 4 reconcile, 0 delete, 0 noop", verKeepOrgOut)

		verOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		respVer := uitest.JSONUIFromBytes(t, []byte(verOut))

		expectedVer := []map[string]string{{
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config-ver-2",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret-ver-2",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, respVer.Tables, expectedVer, "Op:      2 create, 2 delete, 0 update, 0 noop",
			"Wait to: 2 reconcile, 2 delete, 0 noop", verOut)
	})

	cleanUp()

	logger.Section("Scenario-3 [Versioned-keep-original->Non Versioned]", func() {
		verKeepOrgOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		respVerKeepOrg := uitest.JSONUIFromBytes(t, []byte(verKeepOrgOut))

		expectedVerKeepOrg := []map[string]string{
			{
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "ConfigMap",
				"name":            "config-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "Secret",
				"name":            "secret",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"conditions":      "",
				"kind":            "Secret",
				"name":            "secret-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			},
		}

		validateChanges(t, respVerKeepOrg.Tables, expectedVerKeepOrg, "Op:      4 create, 0 delete, 0 update, 0 noop",
			"Wait to: 4 reconcile, 0 delete, 0 noop", verKeepOrgOut)

		nonVerOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		respNonVer := uitest.JSONUIFromBytes(t, []byte(nonVerOut))

		expectedVer := []map[string]string{{
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config-ver-1",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}, {
			"conditions":      "",
			"kind":            "Secret",
			"name":            "secret-ver-1",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}}

		validateChanges(t, respNonVer.Tables, expectedVer, "Op:      0 create, 2 delete, 2 update, 0 noop",
			"Wait to: 2 reconcile, 2 delete, 0 noop", nonVerOut)
	})
}

func TestAdoptionOfResourcesWithVersionedAnn(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kubectl := Kubectl{t, env.Namespace, logger}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
`

	name := "test-adoption-of-res-with-ver-ann"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Kapp should adopt already deployed versioned resource through kubectl", func() {
		out, _ := kubectl.RunWithOpts([]string{"apply", "-f", "-", "-o", "json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		respKubectl := kcv1alpha1.App{}

		err := json.Unmarshal([]byte(out), &respKubectl)
		if err != nil {
			t.Fatalf("Expected to successfully unmarshal: %s", err)
		}

		_, versionedAnnExists := respKubectl.Annotations["kapp.k14s.io/versioned"]
		if respKubectl.Kind != "ConfigMap" || respKubectl.Name != "config" || !versionedAnnExists {
			t.Fatalf("Expected to have versioned ConfigMap resource")
		}

		kappOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		respKapp := uitest.JSONUIFromBytes(t, []byte(kappOut))

		expectedKapp := []map[string]string{{
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"conditions":      "",
			"kind":            "ConfigMap",
			"name":            "config-ver-1",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, respKapp.Tables, expectedKapp, "Op:      1 create, 1 delete, 0 update, 0 noop",
			"Wait to: 1 reconcile, 1 delete, 0 noop", kappOut)
	})
}

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
