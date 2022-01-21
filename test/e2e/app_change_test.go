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

	logger.Section("app change list sort on oldest first", func() {
		out, _ := kapp.RunWithOpts([]string{"app-change", "ls", "-a", name, "--sort", "oldest-first", "--json"}, RunOpts{})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equal(t, 2, len(resp.Tables[0].Rows), "Expected to have 2 app-changes")

		require.Equal(t, "update: Op: 1 create, 0 delete, 0 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[0]["description"], "Expected description to match")
		require.Equal(t, "update: Op: 0 create, 0 delete, 1 update, 0 noop, 0 exists / Wait to: 1 reconcile, 0 delete, 0 noop", resp.Tables[0].Rows[1]["description"], "Expected description to match")
	})
}
