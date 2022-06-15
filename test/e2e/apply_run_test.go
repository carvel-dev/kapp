// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyRun(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	// these key's values are random generated or k8s version based output. Will be used to remove these matching line from output and then match with expected output
	keys := []string{"kapp.k14s.io/app", "creationTimestamp:", "resourceVersion:", "uid:", "selfLink:", "kapp.k14s.io/association"}
	yaml := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: simple-cm
data:
  hello_msg: good-morning-bangalore
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: simple-cm1
data:
  hello_msg: hello
`
	name := "test-apply-run"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}
	cleanUp()
	defer cleanUp()
	logger.Section("creating an app with multiple resources", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--apply-run"},
			RunOpts{StdinReader: strings.NewReader(yaml)})
		expectedOutput := `
# add: configmap/simple-cm (v1) namespace: kapp-test
---
apiVersion: v1
data:
  hello_msg: good-morning-bangalore
kind: ConfigMap
metadata:
  labels:
  name: simple-cm
  namespace: kapp-test
# add: configmap/simple-cm1 (v1) namespace: kapp-test
---
apiVersion: v1
data:
  hello_msg: hello
kind: ConfigMap
metadata:
  labels:
  name: simple-cm1
  namespace: kapp-test
Succeeded
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(keys, out)

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, out, expectedOutput, "output does not match1")

		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml)})
	})

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: simple-cm
data:
  hello_msg: good-morning
`
	logger.Section("update configmap simple-cm and remove configmap simple-cm1", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--apply-run"},
			RunOpts{StdinReader: strings.NewReader(yaml1)})
		expectedOutput := `
# update: configmap/simple-cm (v1) namespace: kapp-test
---
apiVersion: v1
data:
  hello_msg: good-morning
kind: ConfigMap
metadata:
  labels:
  name: simple-cm
  namespace: kapp-test
# delete: configmap/simple-cm1 (v1) namespace: kapp-test
Succeeded
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(keys, out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, out, expectedOutput, "output does not match")

		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml1)})
	})

	yaml2 := `
---
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
  namespace: kapp-test
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
`
	logger.Section("remove configmap simple-cm and add a secret", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--apply-run"},
			RunOpts{StdinReader: strings.NewReader(yaml2)})
		expectedOutput := `
# add: secret/mysecret (v1) namespace: kapp-test
---
apiVersion: v1
data:
  password: <-- value not shown (#1)
  username: <-- value not shown (#2)
kind: Secret
metadata:
  labels:
  name: mysecret
  namespace: kapp-test
# delete: configmap/simple-cm (v1) namespace: kapp-test
Succeeded
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(keys, out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Equal(t, out, expectedOutput, "output does not match")
	})
}

// clearKeys will remove all matching strings in keys from out
func clearKeys(keys []string, out string) string {
	for _, key := range keys {
		key = key + ".*"
		r := regexp.MustCompile(key)
		out = r.ReplaceAllString(out, "")

		//removing all empty lines
		r = regexp.MustCompile(`[ ]*[\n\t]*\n`)
		out = r.ReplaceAllString(out, "\n")
		out = strings.ReplaceAll(out, "\n\n", "\n")
	}
	return out
}
