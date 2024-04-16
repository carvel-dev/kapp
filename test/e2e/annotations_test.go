// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
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
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
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

		validateChanges(t, respNonVer.Tables, expectedNonVer, "Op:      2 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 2 reconcile, 0 delete, 0 noop", nonVerOut)

		verOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		respVer := uitest.JSONUIFromBytes(t, []byte(verOut))

		expectedVer := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "ConfigMap",
			"name":            "config-ver-1",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "Secret",
			"name":            "secret-ver-1",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, respVer.Tables, expectedVer, "Op:      2 create, 2 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 2 reconcile, 2 delete, 0 noop", verOut)

		verKeepOrgOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		respVerKeepOrg := uitest.JSONUIFromBytes(t, []byte(verKeepOrgOut))

		expectedVerKeepOrg := []map[string]string{
			{
				"kind":            "ConfigMap",
				"name":            "config",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"kind":            "ConfigMap",
				"name":            "config-ver-2",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"kind":            "Secret",
				"name":            "secret",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
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

		validateChanges(t, respVerKeepOrg.Tables, expectedVerKeepOrg, "Op:      4 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 4 reconcile, 0 delete, 0 noop", verKeepOrgOut)
	})

	cleanUp()

	logger.Section("Scenario-2 [Versioned-keep-original->Versioned]", func() {
		verKeepOrgOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		respVerKeepOrg := uitest.JSONUIFromBytes(t, []byte(verKeepOrgOut))

		expectedVerKeepOrg := []map[string]string{
			{
				"kind":            "ConfigMap",
				"name":            "config",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"kind":            "ConfigMap",
				"name":            "config-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"kind":            "Secret",
				"name":            "secret",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
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

		validateChanges(t, respVerKeepOrg.Tables, expectedVerKeepOrg, "Op:      4 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 4 reconcile, 0 delete, 0 noop", verKeepOrgOut)

		verOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		respVer := uitest.JSONUIFromBytes(t, []byte(verOut))

		expectedVer := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "ConfigMap",
			"name":            "config-ver-2",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}, {
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "Secret",
			"name":            "secret-ver-2",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, respVer.Tables, expectedVer, "Op:      2 create, 2 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 2 reconcile, 2 delete, 0 noop", verOut)
	})

	cleanUp()

	logger.Section("Scenario-3 [Versioned-keep-original->Non Versioned]", func() {
		verKeepOrgOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		respVerKeepOrg := uitest.JSONUIFromBytes(t, []byte(verKeepOrgOut))

		expectedVerKeepOrg := []map[string]string{
			{
				"kind":            "ConfigMap",
				"name":            "config",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"kind":            "ConfigMap",
				"name":            "config-ver-1",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
				"kind":            "Secret",
				"name":            "secret",
				"namespace":       "kapp-test",
				"op":              "create",
				"op_strategy":     "",
				"reconcile_info":  "",
				"reconcile_state": "",
				"wait_to":         "reconcile",
			}, {
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

		validateChanges(t, respVerKeepOrg.Tables, expectedVerKeepOrg, "Op:      4 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 4 reconcile, 0 delete, 0 noop", verKeepOrgOut)

		nonVerOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		respNonVer := uitest.JSONUIFromBytes(t, []byte(nonVerOut))

		expectedVer := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}, {
			"kind":            "ConfigMap",
			"name":            "config-ver-1",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "Secret",
			"name":            "secret",
			"namespace":       "kapp-test",
			"op":              "update",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "reconcile",
		}, {
			"kind":            "Secret",
			"name":            "secret-ver-1",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}}

		validateChanges(t, respNonVer.Tables, expectedVer, "Op:      0 create, 2 delete, 2 update, 0 noop, 0 exists",
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

		respKubectl := corev1.ConfigMap{}

		err := json.Unmarshal([]byte(out), &respKubectl)
		require.NoErrorf(t, err, "Expected to successfully unmarshal")

		_, versionedAnnExists := respKubectl.Annotations["kapp.k14s.io/versioned"]
		require.Condition(t, func() bool {
			return respKubectl.Kind == "ConfigMap" && respKubectl.Name == "config" && versionedAnnExists
		}, "Expected to have versioned ConfigMap resource")

		kappOut, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})

		respKapp := uitest.JSONUIFromBytes(t, []byte(kappOut))

		expectedKapp := []map[string]string{{
			"kind":            "ConfigMap",
			"name":            "config",
			"namespace":       "kapp-test",
			"op":              "delete",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "ok",
			"wait_to":         "delete",
		}, {
			"kind":            "ConfigMap",
			"name":            "config-ver-1",
			"namespace":       "kapp-test",
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}
		validateChanges(t, respKapp.Tables, expectedKapp, "Op:      1 create, 1 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 1 delete, 0 noop", kappOut)
	})
}

func TestVersionedAnnotation_WithFailedPreviousVersions(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	failYaml := `
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job
  annotations:
    kapp.k14s.io/versioned: ""
spec:
  template:
    metadata:
      name: my-job
    spec:
      restartPolicy: Never
      containers:
        - name: my-job
          image: alpine:3.15.0
          command: [ "sh", "-c", "exit 1" ]
  backoffLimit: 0
`

	successYaml := `
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job
  annotations:
    kapp.k14s.io/versioned: ""
spec:
  template:
    metadata:
      name: my-job
    spec:
      restartPolicy: Never
      containers:
        - name: my-job
          image: alpine:3.15.0
          command: [ "sh", "-c", "exit 0" ]
  backoffLimit: 0
`

	name := "test-ver-ann-with-failed-prev-ver"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy ver-1 of job which is expected to fail", func() {
		out, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(failYaml)})

		require.Error(t, err, "Expected to receive an error")

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expectedOutput := []map[string]string{{
			"kind":            "Job",
			"name":            "my-job-ver-1",
			"namespace":       env.Namespace,
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, resp.Tables, expectedOutput, "Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("deploy with no changes and use --diff-changes", func() {
		out, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(failYaml)})

		require.Error(t, err, "Expected to receive an error")

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		// if no changes are made, then wait for the existing version to reconcile (as it failed previously)
		expectedOutput := []map[string]string{{
			"kind":            "Job",
			"name":            "my-job-ver-1",
			"namespace":       env.Namespace,
			"op":              "",
			"op_strategy":     "",
			"reconcile_info":  "Failed with reason\nBackoffLimitExceeded: Job has\nreached the specified backoff limit",
			"reconcile_state": "fail",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, resp.Tables, expectedOutput, "Op:      0 create, 0 delete, 0 update, 1 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})

	logger.Section("deploy new version with a change", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(successYaml)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		// if a change is made, i.e a new version is created, then kapp should ignore previous failed versions
		expectedOutput := []map[string]string{{
			"kind":            "Job",
			"name":            "my-job-ver-2",
			"namespace":       env.Namespace,
			"op":              "create",
			"op_strategy":     "",
			"reconcile_info":  "",
			"reconcile_state": "",
			"wait_to":         "reconcile",
		}}

		validateChanges(t, resp.Tables, expectedOutput, "Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists",
			"Wait to: 1 reconcile, 0 delete, 0 noop", out)
	})
}
