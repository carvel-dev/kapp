// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestFallbackAllowedNamespaces(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	testNamespace := "test-fallback-allowed-namespace"
	testNamespace2 := "test-fallback-allowed-namespace-2"

	rbac := `
---
apiVersion: v1
kind: Namespace
metadata:
  name: __test-ns__
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: scoped-sa
  namespace: __ns__
---
apiVersion: v1
kind: Secret
metadata:
  name: scoped-sa
  namespace: __ns__
  annotations:
    kubernetes.io/service-account.name: scoped-sa
type: kubernetes.io/service-account-token
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: scoped-role
  namespace: __ns__
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["*"]
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: scoped-role
  namespace: __test-ns__
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["*"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: scoped-role-binding
  namespace: __ns__
subjects:
- kind: ServiceAccount
  name: scoped-sa
  namespace: __ns__
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: scoped-role
  namespace: __ns__
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: scoped-role-binding
  namespace: __test-ns__
subjects:
- kind: ServiceAccount
  name: scoped-sa
  namespace: __ns__
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: scoped-role
  namespace: __ns__
`

	rbac = strings.ReplaceAll(rbac, "__ns__", env.Namespace)
	rbac = strings.ReplaceAll(rbac, "__test-ns__", testNamespace)

	rbacName := "test-e2e-rbac-app"
	scopedContext := "scoped-context"
	scopedUser := "scoped-user"
	appName := "test-fallback-allowed-namespace"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", rbacName})
		kapp.Run([]string{"delete", "-a", appName})
		RemoveClusterResource(t, "ns", testNamespace2, "", kubectl)
	}
	cleanUp()
	defer cleanUp()

	kapp.RunWithOpts([]string{"deploy", "-a", rbacName, "-f", "-"}, RunOpts{StdinReader: strings.NewReader(rbac)})
	cleanUpContext := ScopedContext(t, kubectl, "scoped-sa", scopedContext, scopedUser)
	defer cleanUpContext()

	yaml1 := fmt.Sprintf(`
apiVersion: "v1"
kind: ConfigMap
metadata:
  name: cm-1
  namespace: %s
data:
  foo: bar
---
apiVersion: "v1"
kind: ConfigMap
metadata:
  name: cm-2
  namespace: %s
data:
  foo: bar
---
apiVersion: "v1"
kind: ConfigMap
metadata:
  name: cm-3
  namespace: %s
data:
  foo: bar
`, env.Namespace, testNamespace, testNamespace)

	yaml2 := fmt.Sprintf(`
apiVersion: "v1"
kind: ConfigMap
metadata:
  name: cm-1
  namespace: %s
data:
  foo: bar
---
apiVersion: "v1"
kind: ConfigMap
metadata:
  name: cm-2
  namespace: %s
data:
  foo: bar
`, env.Namespace, testNamespace)

	logger.Section("deploy app using scoped context", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-a", appName, "-f", "-", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{StdinReader: strings.NewReader(yaml1)})

		// Expect pod watching error for the fallback allowed namespaces as listing pods is not allowed.
		require.Contains(t, out, fmt.Sprintf(`Pod watching error: pods is forbidden: User cannot list resource "pods" in API group "" in the namespace(s) "%s", "%s"`,
			env.Namespace, testNamespace))

		NewPresentClusterResource("configmap", "cm-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "cm-2", testNamespace, kubectl)
		NewPresentClusterResource("configmap", "cm-3", testNamespace, kubectl)
	})

	logger.Section("inspect app using scoped context", func() {
		out := kapp.Run([]string{"inspect", "-a", appName, "--json", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)})

		expectedResources := []map[string]string{{
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-1",
			"namespace":       env.Namespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-2",
			"namespace":       testNamespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-3",
			"namespace":       testNamespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equalf(t, expectedResources, replaceAge((resp.Tables[0].Rows)), "Expected resources to match")
	})

	logger.Section("delete one configmap and deploy again using scoped context", func() {
		kapp.RunWithOpts([]string{"deploy", "-a", appName, "-f", "-", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{StdinReader: strings.NewReader(yaml2)})

		NewPresentClusterResource("configmap", "cm-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "cm-2", testNamespace, kubectl)
		NewMissingClusterResource(t, "configmap", "cm-3", testNamespace, kubectl)
	})

	logger.Section("delete app", func() {
		kapp.Run([]string{"delete", "-a", appName, fmt.Sprintf("--kubeconfig-context=%s", scopedContext)})

		NewMissingClusterResource(t, "configmap", "cm-1", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "cm-2", testNamespace, kubectl)
		NewMissingClusterResource(t, "configmap", "cm-3", testNamespace, kubectl)
	})

	logger.Section("deploy app with admin permission but scope-to-fallback-allowed-namespaces", func() {
		kapp.RunWithOpts([]string{"deploy", "-a", appName, "-f", "-", "--dangerous-scope-to-fallback-allowed-namespaces"},
			RunOpts{StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("configmap", "cm-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "cm-2", testNamespace, kubectl)
		NewPresentClusterResource("configmap", "cm-3", testNamespace, kubectl)
	})

	logger.Section("inspect app with scope-to-fallback-allowed-namespaces false", func() {
		const appLabelKey string = "kapp.k14s.io/app"
		NewClusterResource(t, "ns", testNamespace2, "", kubectl)
		NewClusterResource(t, "cm", "cm-4", testNamespace2, kubectl)
		labels := NewPresentClusterResource("cm", "cm-2", testNamespace, kubectl).Labels()
		appLabel := labels[appLabelKey]

		patch := fmt.Sprintf(`[{ "op": "add", "path": "/metadata/labels", "value": {%s: "%s"}}]`, appLabelKey, appLabel)
		PatchClusterResource("cm", "cm-4", testNamespace2, patch, kubectl)

		out := kapp.Run([]string{"inspect", "-a", appName, "--json", "--dangerous-scope-to-fallback-allowed-namespaces=false"})

		// Should get the newly added configmap
		expectedResources := []map[string]string{{
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-1",
			"namespace":       env.Namespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-2",
			"namespace":       testNamespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-3",
			"namespace":       testNamespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-4",
			"namespace":       testNamespace2,
			"owner":           "cluster",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(removeCobraFlagDeprecationWarning(out, "dangerous-scope-to-fallback-allowed-namespaces")))

		require.Equalf(t, expectedResources, replaceAge((resp.Tables[0].Rows)), "Expected resources to match")
	})

	logger.Section("inspect app without scope-to-fallback-allowed-namespaces", func() {
		out := kapp.Run([]string{"inspect", "-a", appName, "--json"})

		// Shouldn't get the newly added configmap
		expectedResources := []map[string]string{{
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-1",
			"namespace":       env.Namespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-2",
			"namespace":       testNamespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}, {
			"age":             "<replaced>",
			"kind":            "ConfigMap",
			"name":            "cm-3",
			"namespace":       testNamespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equalf(t, expectedResources, replaceAge((resp.Tables[0].Rows)), "Expected resources to match")
	})

	logger.Section("delete one configmap and deploy again without", func() {
		kapp.RunWithOpts([]string{"deploy", "-a", appName, "-f", "-"},
			RunOpts{StdinReader: strings.NewReader(yaml2)})

		NewPresentClusterResource("configmap", "cm-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "cm-2", testNamespace, kubectl)
		NewMissingClusterResource(t, "configmap", "cm-3", testNamespace, kubectl)
	})

	logger.Section("delete app", func() {
		kapp.Run([]string{"delete", "-a", appName})

		NewMissingClusterResource(t, "configmap", "cm-1", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "cm-2", testNamespace, kubectl)
		NewMissingClusterResource(t, "configmap", "cm-3", testNamespace, kubectl)
	})
}

func ScopedContext(t *testing.T, kubectl Kubectl, _, contextName, userName string) func() {
	token := kubectl.Run([]string{"get", "secret", "scoped-sa", "-o", "jsonpath={.data.token}"})

	tokenDecoded, err := base64.StdEncoding.DecodeString(token)
	require.NoError(t, err)

	currentContextCluster := kubectl.Run([]string{"config", "view", "--minify", "-o", "jsonpath={.clusters[].name}"})

	kubectl.RunWithOpts([]string{"config", "set-credentials", userName, fmt.Sprintf("--token=%s", string(tokenDecoded))},
		RunOpts{NoNamespace: true, Redact: true})

	kubectl.RunWithOpts([]string{"config", "set-context", contextName, fmt.Sprintf("--user=%s", userName), fmt.Sprintf("--cluster=%s", currentContextCluster)},
		RunOpts{NoNamespace: true})

	return func() {
		kubectl.Run([]string{"config", "delete-context", contextName})
		kubectl.Run([]string{"config", "delete-user", userName})
	}
}

func removeCobraFlagDeprecationWarning(output, flag string) string {
	newOut := ""
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("Flag --%s has been deprecated", flag)) {
			continue
		}
		newOut += line + "\n"
	}
	return newOut
}
