// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
	"time"
)

func TestExistsOpWithPlaceholderAnn(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

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
  namespace: kapp-ns
  annotations:
    kapp.k14s.io/external-resource: ""
`

	name := "app"
	externalResourceName := "external"
	externalResourceNamespace := "kapp-ns"
	externalResourceKind := "configmap"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		RemoveClusterResource(t, externalResourceKind, externalResourceName, externalResourceNamespace, kubectl)
	}
	cleanUp()
	defer cleanUp()

	go func() {
		time.Sleep(2 * time.Second)
		NewClusterResource(t, externalResourceKind, externalResourceName, externalResourceNamespace, kubectl)
	}()

	logger.Section("deploying app with external resource annotation", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(app)})

		expectedOutput := `
Changes

Namespace  Name      Kind       Conds.  Age  Op      Op st.  Wait to    Rs  Ri  $
(cluster)  kapp-ns   Namespace  -       -    create  -       reconcile  -   -  $
kapp-ns    external  ConfigMap  -       -    exists  -       reconcile  -   -  $

Op:      1 create, 0 delete, 0 update, 0 noop
Wait to: 2 reconcile, 0 delete, 0 noop

<replaced>: ---- applying 1 changes [0/2 done] ----
<replaced>: create namespace/kapp-ns (v1) cluster
<replaced>: ---- waiting on 1 changes [0/2 done] ----
<replaced>: ok: reconcile namespace/kapp-ns (v1) cluster
<replaced>: ---- applying 1 changes [1/2 done] ----
<replaced>: exists configmap/external (v1) namespace: kapp-ns
<replaced>:  ^ Retryable error: External resource doesn't exists
<replaced>: exists configmap/external (v1) namespace: kapp-ns
<replaced>: ---- waiting on 1 changes [1/2 done] ----
<replaced>: ok: reconcile configmap/external (v1) namespace: kapp-ns
<replaced>: ---- applying complete [2/2 done] ----
<replaced>: ---- waiting complete [2/2 done] ----

Succeeded`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		if expectedOutput != out {
			t.Fatalf("Expected output to be >>%s<<, but got >>%s<<\n", expectedOutput, out)
		}
	})

	logger.Section("inspecting app", func() {
		out, _ := kapp.RunWithOpts([]string{"inspect", "-a", name},
			RunOpts{})
		out = strings.TrimSpace(replaceTarget(replaceAgeStr(replaceSpaces(replaceTs(out)))))

		expectedOutput := `
Resources in app 'app'

Namespace  Name     Kind       Owner  Conds.  Rs  Ri  Age  $
(cluster)  kapp-ns  Namespace  kapp   -       ok  -   2s  $

Rs: Reconcile state
Ri: Reconcile information

1 resources

Succeeded`

		expectedOutput = strings.TrimSpace(replaceAgeStr(replaceSpaces(expectedOutput)))
		if expectedOutput != out {
			t.Fatalf("Expected output to be >>%s<<, but got >>%s<<\n", expectedOutput, out)
		}
	})
}
