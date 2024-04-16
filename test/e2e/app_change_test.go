// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"carvel.dev/kapp/pkg/kapp/app"
	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestAppChange(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: %s
`

	name := "test-app-change"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy app", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, "backend"))})
	})

	logger.Section("deploy with changes", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, "frontend"))})
	})

	logger.Section("app change list", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 2, len(resp.Tables[0].Rows), "Expected to have 2 app-changes")
		require.Equal(t, "update: Op: 0 create, 0 delete, 1 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[0]["description"], "Expected description to match")
		require.Equal(t, "update: Op: 1 create, 0 delete, 0 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[1]["description"], "Expected description to match")
	})

	logger.Section("app change list filter with before flag", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--before", time.Now().Add(1 * time.Second).Format(time.RFC3339), "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 2, len(resp.Tables[0].Rows), "Expected to have 2 app-changes")
		require.Equal(t, "update: Op: 0 create, 0 delete, 1 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[0]["description"], "Expected description to match")
		require.Equal(t, "update: Op: 1 create, 0 delete, 0 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[1]["description"], "Expected description to match")

		out2, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--before", time.Now().Add(-1 * time.Minute).Format(time.RFC3339), "--json"}, RunOpts{})

		resp2 := uitest.JSONUIFromBytes(t, []byte(out2))

		require.Equal(t, 0, len(resp2.Tables[0].Rows), "Expected to have 0 app-changes")
	})

	logger.Section("app change list filter with after flag", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--after", time.Now().Add(24 * time.Hour).Format("2006-01-02"), "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 0, len(resp.Tables[0].Rows), "Expected to have 0 app-changes")

		out2, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--after", time.Now().Add(-1 * time.Minute).Format("2006-01-02"), "--json"}, RunOpts{})

		resp2 := uitest.JSONUIFromBytes(t, []byte(out2))

		require.Equal(t, 2, len(resp2.Tables[0].Rows), "Expected to have 2 app-changes")
	})
}

func TestAppChangeWithLongAppName(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: %s
`

	name := "test-app-change-with-a-very-very-very-very-very-very-very-long-app-name"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy app", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, "backend"))})
	})

	logger.Section("deploy with changes", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, "frontend"))})
	})

	logger.Section("app change list", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 2, len(resp.Tables[0].Rows), "Expected to have 2 app-changes")
		require.Equal(t, "update: Op: 0 create, 0 delete, 1 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[0]["description"], "Expected description to match")
		require.Equal(t, "update: Op: 1 create, 0 delete, 0 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[1]["description"], "Expected description to match")
	})
}

func TestAppChangesMaxToKeep(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, Logger{}}

	rbacName := "test-e2e-rbac-app"
	scopedContext := "scoped-context"
	scopedUser := "scoped-user"
	name := "test-app-changes-max-to-keep"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", rbacName})
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	rbac := fmt.Sprintf(`
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
  verbs: ["get", "list", "watch", "patch", "update", "create"] # no delete permission
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["%s", "%s"]
  verbs: ["delete"] # delete permission for meta configmap only
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["*"]
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
`, name, name+app.AppSuffix)

	kapp.RunWithOpts([]string{"deploy", "-a", rbacName, "-f", "-"}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(rbac)})
	cleanUpContext := ScopedContext(t, kubectl, "scoped-sa", scopedContext, scopedUser)
	defer cleanUpContext()

	yaml1 := `
---
apiVersion: v1
kind: Secret
metadata:
  name: redis-config
`
	logger.Section("Setting app-changes-max-to-keep to 0 doesn't create new app-changes", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--app-changes-max-to-keep=0", fmt.Sprintf("--kubeconfig-context=%s", scopedContext)},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		out, _ = kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 0, len(resp.Tables[0].Rows), "Expected to have 0 app changes")

		out = kapp.Run([]string{"delete", "-a", name, fmt.Sprintf("--kubeconfig-context=%s", scopedContext)})
	})

}

type usedGK struct {
	Group string `yaml:"Group"`
	Kind  string `yaml:"Kind"`
}

type lastChange struct {
	Namespaces []string `yaml:"namespaces"`
}

type yamlSubset struct {
	LastChange lastChange `yaml:"lastChange"`
	UsedGKs    []usedGK   `yaml:"usedGKs"`
}
