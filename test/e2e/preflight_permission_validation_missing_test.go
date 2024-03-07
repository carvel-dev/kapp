// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreflightPermissionValidationMissingPermissions(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	testName := "preflight-permission-validation-missing-permissions"

	base := `
---
apiVersion: v1
kind: Namespace
metadata:
  name: __test-name__
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
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: __test-name__
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["list"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["roles", "rolebindings"]
  verbs: ["get", "list"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: __test-name__
subjects:
- kind: ServiceAccount
  name: scoped-sa
  namespace: __ns__
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: __test-name__
`

	base = strings.ReplaceAll(base, "__test-name__", testName)
	base = strings.ReplaceAll(base, "__ns__", env.Namespace)
	baseName := "preflight-permission-validation-base-app"
	appName := "preflight-permission-validation-app"
	scopedContext := "scoped-context"
	scopedUser := "scoped-user"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", baseName})
		kapp.Run([]string{"delete", "-a", appName})
		RemoveClusterResource(t, "ns", testName, "", kubectl)
	}
	cleanUp()
	defer cleanUp()

	kapp.RunWithOpts([]string{"deploy", "-a", baseName, "-f", "-"}, RunOpts{StdinReader: strings.NewReader(base)})
	cleanUpContext := ScopedContext(t, kubectl, testName, scopedContext, scopedUser)
	defer cleanUpContext()

	basicResource := `
---
apiVersion: v1
kind: Pod
metadata:
  name: __test-name__
  namespace: __test-name__
spec:
  containers:
  - name: simple-app
    image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
    env:
    - name: HELLO_MSG
      value: stranger
`
	basicResource = strings.ReplaceAll(basicResource, "__test-name__", testName)
	logger.Section("attempt to deploy app with a Pod and missing permissions to create Pods", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "--preflight=PermissionValidation", "-a", appName, "-f", "-", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{StdinReader: strings.NewReader(basicResource), AllowError: true})

		require.Error(t, err)
		require.Contains(t, err.Error(), "running preflight check \"PermissionValidation\": not permitted to \"create\" /v1, Resource=pods")
		NewMissingClusterResource(t, "pod", testName, testName, kubectl)
	})

	roleResource := `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: __test-name__
  name: __test-name__
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["create", "update", "delete"]
`

	roleResource = strings.ReplaceAll(roleResource, "__test-name__", testName)
	logger.Section("attempt to deploy app with a Role and missing permissions to create Roles", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "--preflight=PermissionValidation", "-a", appName, "-f", "-", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{StdinReader: strings.NewReader(roleResource), AllowError: true})

		require.Error(t, err)
		require.Contains(t, err.Error(), "running preflight check \"PermissionValidation\": not permitted to \"create\" rbac.authorization.k8s.io/v1, Resource=roles")
		NewMissingClusterResource(t, "role", testName, testName, kubectl)
	})

	bindingResource := `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: __test-name__
  name: __test-name__
subjects:
  - kind: ServiceAccount
    namespace: __test-name__
    name: default
roleRef:
  kind: ClusterRole
  name: admin
  apiGroup: rbac.authorization.k8s.io
`
	bindingResource = strings.ReplaceAll(bindingResource, "__test-name__", testName)
	logger.Section("attempt to deploy app with a RoleBinding and missing permissions to create RoleBindings", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "--preflight=PermissionValidation", "-a", appName, "-f", "-", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{StdinReader: strings.NewReader(bindingResource), AllowError: true})

		require.Error(t, err)
		require.Contains(t, err.Error(), "running preflight check \"PermissionValidation\": not permitted to \"create\" rbac.authorization.k8s.io/v1, Resource=rolebindings")
		NewMissingClusterResource(t, "rolebinding", testName, testName, kubectl)
	})
}
