// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExistsAnn(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	namespace := `
apiVersion: v1
kind: Namespace
metadata:
  name: external
`
	app := `
apiVersion: v1
kind: Namespace
metadata:
  name: external
  annotations:
    kapp.k14s.io/exists: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kapp-config
  namespace: external
`

	name := "app"
	externalApp := "external-app"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", externalApp})
	}
	cleanUp()
	defer cleanUp()

	go func() {
		time.Sleep(2 * time.Second)
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", externalApp},
			RunOpts{StdinReader: strings.NewReader(namespace)})
	}()

	logger.Section("deploying app with exists annotation for a non existing resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})

		expectedOutput := `
Changes

Namespace  Name         Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
(cluster)  external     Namespace  -    exists  -       reconcile  -   -  $
external   kapp-config  ConfigMap  -    create  -       reconcile  -   -  $

Op:      1 create, 0 delete, 0 update, 0 noop, 1 exists
Wait to: 2 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/2 done] ----
<replaced>: exists namespace/external (v1) cluster
<replaced>:  ^ Retryable error: External resource doesn't exists
<replaced>: exists namespace/external (v1) cluster
<replaced>: ---- waiting on 1 changes [0/2 done] ----
<replaced>: ok: reconcile namespace/external (v1) cluster
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: create configmap/kapp-config (v1) namespace: external
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: reconcile configmap/kapp-config (v1) namespace: external
<replaced>: ---- applying complete [2/2 done] ----
<replaced>: ---- waiting complete [2/2 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, expectedOutput, out)
	})

	logger.Section("deploying app with exists annotation for an already existing resource", func() {
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

Namespace  Name         Kind       Owner  Rs  Ri  Age  $
external   kapp-config  ConfigMap  kapp   ok  -   <replaced>  $

Rs: Reconcile state
Ri: Reconcile information

1 resources

Succeeded`

		expectedOutput = strings.TrimSpace(replaceAgeStr(replaceSpaces(expectedOutput)))
		require.Equal(t, expectedOutput, out)
	})
}
