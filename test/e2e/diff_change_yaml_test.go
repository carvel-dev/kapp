// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffChangeYAML(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	fieldsExcludedInMatch := []string{"kapp.k14s.io/app", "creationTimestamp:", "resourceVersion:", "uid:", "selfLink:", "kapp.k14s.io/association"}
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
	name := "test-changes-yaml"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("creating an app with multiple resources", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes-yaml"},
			RunOpts{StdinReader: strings.NewReader(yaml)})
		expectedOutput := `
---
# create: configmap/simple-cm (v1) namespace: kapp-test
apiVersion: v1
data:
  hello_msg: good-morning-bangalore
kind: ConfigMap
metadata:
  labels:
  name: simple-cm
  namespace: kapp-test
---
# create: configmap/simple-cm1 (v1) namespace: kapp-test
apiVersion: v1
data:
  hello_msg: hello
kind: ConfigMap
metadata:
  labels:
  name: simple-cm1
  namespace: kapp-test
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(fieldsExcludedInMatch, out)

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")
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

		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes-yaml"},
			RunOpts{StdinReader: strings.NewReader(yaml1)})
		expectedOutput := `
---
# update: configmap/simple-cm (v1) namespace: kapp-test
apiVersion: v1
data:
  hello_msg: good-morning
kind: ConfigMap
metadata:
  labels:
  name: simple-cm
  namespace: kapp-test
---
# delete: configmap/simple-cm1 (v1) namespace: kapp-test
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(fieldsExcludedInMatch, out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")
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
		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml1)})
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes-yaml"},
			RunOpts{StdinReader: strings.NewReader(yaml2)})
		expectedOutput := `
---
# create: secret/mysecret (v1) namespace: kapp-test
apiVersion: v1
data:
  password: <-- value not shown (#1)
  username: <-- value not shown (#2)
kind: Secret
metadata:
  labels:
  name: mysecret
  namespace: kapp-test
---
# delete: configmap/simple-cm (v1) namespace: kapp-test
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = clearKeys(fieldsExcludedInMatch, out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")
	})
}

// clearKeys will remove all matching fields in fieldsExcludedInMatch from out
func clearKeys(fieldsExcludedInMatch []string, out string) string {
	for _, field := range fieldsExcludedInMatch {
		r := regexp.MustCompile(field + ".*")
		out = r.ReplaceAllString(out, "")
	}

	// removing all empty lines, extra tab or space
	r1 := regexp.MustCompile(`[ ]*[\n\t]*\n`)
	out = r1.ReplaceAllString(out, "\n")
	r1 = regexp.MustCompile(`[\n]+`)
	out = r1.ReplaceAllString(out, "\n")
	return out
}
