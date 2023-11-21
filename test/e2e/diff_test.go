// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
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
			"kind":            "ConfigMap",
			"name":            "redis-config2",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "",
		}}

		require.Exactlyf(t, expected, resp.Tables[0].Rows, "Expected to see correct changes, but did not")
		require.Equalf(t, "Op:      3 create, 0 delete, 0 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
		require.Equalf(t, "Wait to: 3 reconcile, 0 delete, 0 noop", resp.Tables[0].Notes[1], "Expected to see correct summary, but did not")
	})

	logger.Section("deploy no change", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))
		expected := []map[string]string{}

		require.Exactlyf(t, expected, resp.Tables[0].Rows, "Expected to see correct changes, but did not")
		require.Equalf(t, "Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
		require.Equalf(t, "Wait to: 0 reconcile, 0 delete, 0 noop", resp.Tables[0].Notes[1], "Expected to see correct summary, but did not")
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
			"kind":            "ConfigMap",
			"name":            "redis-config3",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "",
		}}

		require.Exactlyf(t, expected, replaceAge(resp.Tables[0].Rows), "Expected to see correct changes, but did not")
		require.Equalf(t, "Op:      1 create, 1 delete, 1 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
		require.Equalf(t, "Wait to: 2 reconcile, 1 delete, 0 noop", resp.Tables[0].Notes[1], "Expected to see correct summary, but did not")
	})

	logger.Section("delete", func() {
		out, _ := kapp.RunWithOpts([]string{"delete", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		expected := []map[string]string{{
			"age":             "<replaced>",
			"op":              "delete",
			"op_strategy":     "",
			"wait_to":         "delete",
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
			"kind":            "ConfigMap",
			"name":            "redis-config3",
			"namespace":       "kapp-test",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		require.Exactlyf(t, expected, replaceAge(resp.Tables[0].Rows), "Expected to see correct changes, but did not")
		require.Equalf(t, "Op:      0 create, 3 delete, 0 update, 0 noop, 0 exists", resp.Tables[0].Notes[0], "Expected to see correct summary, but did not")
		require.Equalf(t, "Wait to: 0 reconcile, 3 delete, 0 noop", resp.Tables[0].Notes[1], "Expected to see correct summary, but did not")
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

	require.Errorf(t, err, "Expected to receive error")

	require.Containsf(t, err.Error(), "Exiting after diffing with no pending changes (exit status 2)", "Expected to find stderr output")
	require.Containsf(t, err.Error(), "exit code: '2'", "Expected to find exit code")

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
`

	_, err = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
		"--diff-run", "--diff-exit-status"},
		RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

	require.Errorf(t, err, "Expected to receive error")

	require.Containsf(t, err.Error(), "Exiting after diffing with pending changes (exit status 3)", "Expected to find stderr output")
	require.Containsf(t, err.Error(), "exit code: '3'", "Expected to find exit code")
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

	require.Containsf(t, out, expectedOutput, "Did not find expected diff output")

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

	require.Containsf(t, out, expectedOutput, "Did not find expected diff output")
}

func TestDiffRun(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}

	name := "not-create-configmap-diff-run"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
`

	_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name,
		"--diff-run"},
		RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

	NewMissingClusterResource(t, "configmap", name, env.Namespace, kubectl)
}

func TestDiffRun_WithReadOnlyPermissions(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}
	kubectl := Kubectl{t, env.Namespace, Logger{}}

	rbac := `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: scoped-sa
---
apiVersion: v1
kind: Secret
metadata:
  name: scoped-sa
  annotations:
    kubernetes.io/service-account.name: scoped-sa
type: kubernetes.io/service-account-token
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: scoped-role
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: scoped-role-binding
subjects:
- kind: ServiceAccount
  name: scoped-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: scoped-role
`

	rbacName := "test-e2e-rbac-app"
	scopedContext := "scoped-context"
	scopedUser := "scoped-user"
	appName := "diff-run-read-only"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", rbacName})
		kapp.Run([]string{"delete", "-a", appName})
	}
	cleanUp()
	defer cleanUp()

	kapp.RunWithOpts([]string{"deploy", "-a", rbacName, "-f", "-"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(rbac)})
	cleanUpContext := ScopedContext(t, kubectl, "scoped-sa", scopedContext, scopedUser)
	defer cleanUpContext()

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
`

	logger.Section("diff-run app create using read-only role", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--diff-run", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewMissingClusterResource(t, "configmap", appName, env.Namespace, kubectl)
	})

	logger.Section("diff-run app update using read-only role", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("configmap", appName, env.Namespace, kubectl)

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--diff-run", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})
}

func TestAnchoredDiff(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	name := "test-anchored-diff"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	yaml1 := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 9.5.5
             url: docker.io/grafana/grafana:9.5.5
        url: index.docker.io/grafana/grafana@sha256:6c6fe32401b6b14e1886e61a7bacd5cc4b6fbd0de1e58e985db0e48f99fe1be1
      - origins:
        - resolved:
            tag: 1.24.3
            url: quay.io/kiwigrid/k8s-sidecar:1.24.3
        url: quay.io/kiwigrid/k8s-sidecar@sha256:5af76eebbba79edf4f7471bf1c3d5f2b40858114730c92d95eafe5716abe1fe8
data:

`

	yaml2 := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 10.1.4
            url: docker.io/grafana/grafana:10.1.4
        url: index.docker.io/grafana/grafana@sha256:29f39e23705d3ef653fa84ca3c01731e0771f1fedbd69ecb99868270cdeb0572
      - origins:
        - resolved:
            tag: 1.25.1
            url: quay.io/kiwigrid/k8s-sidecar:1.25.1
        url: quay.io/kiwigrid/k8s-sidecar@sha256:415d07ee1027c3ff7af9e26e05e03ffd0ec0ccf9f619ac00ab24366efe4343bd
data:

`
	// Add keys so that number of lines in the yamls are > 500
	for i := 0; i <= 500; i++ {
		line := fmt.Sprintf("  key%v: value%v\n", i, i)
		yaml1 += line
		yaml2 += line
	}

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy without anchored diff", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--diff-run", "--diff-summary=false"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		expectedDiff := `
@@ update configmap/cm-1 (v1) namespace: kapp-test @@
  ...
508,508           - resolved:
509     -             tag: 9.5.5
510     -              url: docker.io/grafana/grafana:9.5.5
511     -         url: index.docker.io/grafana/grafana@sha256:6c6fe32401b6b14e1886e61a7bacd5cc4b6fbd0de1e58e985db0e48f99fe1be1
    509 +             tag: 10.1.4
    510 +             url: docker.io/grafana/grafana:10.1.4
    511 +         url: index.docker.io/grafana/grafana@sha256:29f39e23705d3ef653fa84ca3c01731e0771f1fedbd69ecb99868270cdeb0572
512,512         - origins:
513,513           - resolved:
514     -             tag: 1.24.3
515     -             url: quay.io/kiwigrid/k8s-sidecar:1.24.3
516     -         url: quay.io/kiwigrid/k8s-sidecar@sha256:5af76eebbba79edf4f7471bf1c3d5f2b40858114730c92d95eafe5716abe1fe8
    514 +             tag: 1.25.1
    515 +             url: quay.io/kiwigrid/k8s-sidecar:1.25.1
    516 +         url: quay.io/kiwigrid/k8s-sidecar@sha256:415d07ee1027c3ff7af9e26e05e03ffd0ec0ccf9f619ac00ab24366efe4343bd
517,517     creationTimestamp: "2006-01-02T15:04:05Z07:00"
518,518     labels:

Succeeded
`
		require.Equal(t, expectedDiff, replaceTimestampWithDfaultValue(replaceTarget(out)))
	})

	logger.Section("deploy with anchored diff", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--diff-run", "--diff-summary=false", "--diff-anchored"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		expectedDiff := `
@@ update configmap/cm-1 (v1) namespace: kapp-test @@
  ...
508,508           - resolved:
509     -             tag: 9.5.5
510     -              url: docker.io/grafana/grafana:9.5.5
511     -         url: index.docker.io/grafana/grafana@sha256:6c6fe32401b6b14e1886e61a7bacd5cc4b6fbd0de1e58e985db0e48f99fe1be1
512     -       - origins:
513     -         - resolved:
514     -             tag: 1.24.3
515     -             url: quay.io/kiwigrid/k8s-sidecar:1.24.3
516     -         url: quay.io/kiwigrid/k8s-sidecar@sha256:5af76eebbba79edf4f7471bf1c3d5f2b40858114730c92d95eafe5716abe1fe8
    509 +             tag: 10.1.4
    510 +             url: docker.io/grafana/grafana:10.1.4
    511 +         url: index.docker.io/grafana/grafana@sha256:29f39e23705d3ef653fa84ca3c01731e0771f1fedbd69ecb99868270cdeb0572
    512 +       - origins:
    513 +         - resolved:
    514 +             tag: 1.25.1
    515 +             url: quay.io/kiwigrid/k8s-sidecar:1.25.1
    516 +         url: quay.io/kiwigrid/k8s-sidecar@sha256:415d07ee1027c3ff7af9e26e05e03ffd0ec0ccf9f619ac00ab24366efe4343bd
517,517     creationTimestamp: "2006-01-02T15:04:05Z07:00"
518,518     labels:

Succeeded
`
		require.Equal(t, expectedDiff, replaceTimestampWithDfaultValue(replaceTarget(out)))
	})
}
