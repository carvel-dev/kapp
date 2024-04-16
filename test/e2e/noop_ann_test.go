// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopAnn(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	configMap := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: external
`

	app := `
apiVersion: v1
kind: Namespace
metadata:
  name: kapp-ns
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: external
  annotations:
    kapp.k14s.io/noop: ""
`

	name := "app"
	externalApp := "external-app"

	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", externalApp})
	}
	cleanUp()
	defer cleanUp()

	logger.Section("deploying app with noop annotation for a non existing resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})

		expectedOutput := `
Changes

Namespace  Name     Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
(cluster)  kapp-ns  Namespace  -    create  -       reconcile  -   -  $

Op:      1 create, 0 delete, 0 update, 0 noop, 0 exists
Wait to: 1 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/1 done] ----
<replaced>: create namespace/kapp-ns (v1) cluster
<replaced>: ---- waiting on 1 changes [0/1 done] ----
<replaced>: ok: reconcile namespace/kapp-ns (v1) cluster
<replaced>: ---- applying complete [1/1 done] ----
<replaced>: ---- waiting complete [1/1 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, expectedOutput, out)
	})

	logger.Section("deploying app with noop annotation for an existing resource", func() {
		// deploying app with the external resource
		kapp.RunWithOpts([]string{"deploy", "-a", externalApp, "-f", "-"},
			RunOpts{StdinReader: strings.NewReader(configMap)})

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})

		expectedOutput := `
Changes

Namespace  Name  Kind  Age  Op  Op st.  Wait to  Rs  Ri  $

Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists
Wait to: 0 reconcile, 0 delete, 0 noop

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, expectedOutput, out)
	})

	logger.Section("inspecting app", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name, "--tty=true"},
			RunOpts{})
		out = strings.TrimSpace(replaceTarget(replaceAgeStr(replaceSpaces(replaceTs(out)))))

		expectedOutput := `
Resources in app 'app'

Namespace  Name     Kind       Owner  Rs  Ri  Age  $
(cluster)  kapp-ns  Namespace  kapp   ok  -   2s  $

Rs: Reconcile state
Ri: Reconcile information

1 resources

Succeeded`

		expectedOutput = strings.TrimSpace(replaceAgeStr(replaceSpaces(expectedOutput)))
		require.Equal(t, expectedOutput, out)
	})
}
