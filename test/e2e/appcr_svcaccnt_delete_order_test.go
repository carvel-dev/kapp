// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestAppCRSvcAccntDeleteOrder(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml := `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default-ns-sa
  namespace: kapp-test
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: default-ns-role
  namespace: kapp-test
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: default-ns-role-binding
  namespace: kapp-test
subjects:
- kind: ServiceAccount
  name: default-ns-sa
  namespace: kapp-test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: default-ns-role
---
apiVersion: kappctrl.k14s.io/v1alpha1
kind: App
metadata:
  name: simple-app-cr
  namespace: kapp-test
spec:
  serviceAccountName: default-ns-sa
  fetch:
  - git:
      url: https://github.com/k14s/k8s-simple-app-example
      ref: origin/develop
      subPath: config-step-2-template
  template:
  - ytt: {}
  deploy:
  - kapp: {}
`
	name := "test-appcr-svcaccnt-delete-order"
	kcappname := "kc"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", kcappname})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("install kapp-controller on cluster", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "https://github.com/vmware-tanzu/carvel-kapp-controller/releases/latest/download/release.yml", "-a", kcappname},
			RunOpts{})
		// TODO: how to test presence of kc
	})

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		NewPresentClusterResource("app", "simple-app-cr", env.Namespace, kubectl)
		NewPresentClusterResource("role", "default-ns-role", env.Namespace, kubectl)
		NewPresentClusterResource("rolebinding", "default-ns-role-binding", env.Namespace, kubectl)
		NewPresentClusterResource("serviceaccount", "default-ns-sa", env.Namespace, kubectl)
	})

	logger.Section("delete app", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})

		NewMissingClusterResource(t, "app", "simple-app-cr", env.Namespace, kubectl)
		NewMissingClusterResource(t, "role", "default-ns-role", env.Namespace, kubectl)
		NewMissingClusterResource(t, "rolebinding", "default-ns-role-binding", env.Namespace, kubectl)
		NewMissingClusterResource(t, "serviceaccount", "default-ns-sa", env.Namespace, kubectl)
	})
}
