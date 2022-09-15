// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeriodicUpdate(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
data:
  hello_msg: carvel
kind: ConfigMap
metadata:
  annotations:
    kapp.k14s.io/max-duration: 2s
  name: simple-cm1
---
apiVersion: v1
data: 
  hello_msg: carvel
kind: ConfigMap
metadata: 
  annotations: 
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/max-duration: 2s
  name: simple-cm2
`

	yaml2 := `
--- 
apiVersion: v1
data:
  hello_msg: carvel
kind: ConfigMap
metadata:
  name: simple-cm1
---
apiVersion: v1
data: 
  hello_msg: carvel
kind: ConfigMap
metadata: 
  annotations: 
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/max-duration: 2s
  name: simple-cm2
`

	name := "test-periodic-update"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{StdinReader: strings.NewReader(yaml1)})
		NewPresentClusterResource("ConfigMap", "simple-cm1", env.Namespace, kubectl)
		NewPresentClusterResource("ConfigMap", "simple-cm2-ver-1", env.Namespace, kubectl)
	})

	logger.Section("Deploy again before max-duration expired", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--diff-run", "--diff-summary=false", "--tty=false"}, RunOpts{StdinReader: strings.NewReader(yaml1)})
		require.Equal(t, "", out, "output does not match")
	})

	time.Sleep(2 * time.Second)
	logger.Section("Deploy again after 2 second", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--diff-run", "--diff-summary=false"}, RunOpts{StdinReader: strings.NewReader(yaml1)})

		expectedOutput := `
@@ create configmap/simple-cm2-ver-2 (v1) namespace: kapp-test @@
  ...
  5,  5     annotations:
      6 +     kapp.k14s.io/last-renewed-time: "2006-01-02T15:04:05Z07:00"
  6,  7       kapp.k14s.io/max-duration: 2s
  7,  8       kapp.k14s.io/versioned: ""
@@ update configmap/simple-cm1 (v1) namespace: kapp-test @@
  ...
  5,  5     annotations:
      6 +     kapp.k14s.io/last-renewed-time: "2006-01-02T15:04:05Z07:00"
  6,  7       kapp.k14s.io/max-duration: 2s
  7,  8     creationTimestamp: "2006-01-02T15:04:05Z07:00"
`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = replaceTimestampWithDfaultValue(out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")
	})

	time.Sleep(2 * time.Second)
	logger.Section("Deploy again after 2 second", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run"}, RunOpts{StdinReader: strings.NewReader(yaml2)})

		expectedOutput := `
Changes

Namespace  Name              Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
kapp-test  simple-cm1        ConfigMap  4s   update  -       reconcile  ok  -  $
^          simple-cm2-ver-2  ConfigMap  -    create  -       reconcile  -   -  $

Op:      1 create, 0 delete, 1 update, 0 noop, 0 exists
Wait to: 2 reconcile, 0 delete, 0 noop

Succeeded
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		r := regexp.MustCompile("[0-9]s")
		out = r.ReplaceAllString(out, "4s")

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, expectedOutput, out, "output does not match")
	})

	kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{StdinReader: strings.NewReader(yaml2)})

	time.Sleep(2 * time.Second)
	logger.Section("Deploy again after 2 second", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-run"}, RunOpts{StdinReader: strings.NewReader(yaml1)})

		expectedOutput := `
Changes

Namespace  Name              Kind       Age  Op      Op st.  Wait to    Rs  Ri  $
kapp-test  simple-cm1        ConfigMap  6s   update  -       reconcile  ok  -  $
^          simple-cm2-ver-3  ConfigMap  -    create  -       reconcile  -   -  $

Op:      1 create, 0 delete, 1 update, 0 noop, 0 exists
Wait to: 2 reconcile, 0 delete, 0 noop

Succeeded
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		r := regexp.MustCompile("[0-9]s")
		out = r.ReplaceAllString(out, "6s")

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, expectedOutput, out, "output does not match")
	})
}

func replaceTimestampWithDfaultValue(out string) string {
	r := regexp.MustCompile("[1-9]\\d{3}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z")
	out = r.ReplaceAllString(out, time.RFC3339)
	return out
}
