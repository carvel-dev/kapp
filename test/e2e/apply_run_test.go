// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyRun(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
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
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})
		expectedOutput := `
# add: configmap/simple-cm (v1) namespace: kapp-test
---
apiVersion: v1
data:
  hello_msg: good-morning-bangalore
kind: ConfigMap
metadata:
  labels:
    kapp.k14s.io/app: "1654768619446772000"
    kapp.k14s.io/association: v1.aa6ba70e7f14b3140de8009fda0a6fad
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
    kapp.k14s.io/app: "1654768619446772000"
    kapp.k14s.io/association: v1.df359155db7824da8b7d86ec097a40cf
  name: simple-cm1
  namespace: kapp-test

Succeeded
`
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		// filtering value for label "kapp.k14s.io/app:" and updating with hardcode value
		if out != "" && len(strings.Split(out, "kapp.k14s.io/app: ")) > 2 {
			repValue1 := strings.Split((strings.Split(out, "kapp.k14s.io/app: ")[1]), "\n")[0]
			out = strings.Replace(out, repValue1, `"1654768619446772000"`, 1)
			repValue2 := strings.Split((strings.Split(out, "kapp.k14s.io/app: ")[2]), "\n")[0]
			out = strings.Replace(out, repValue2, `"1654768619446772000"`, 1)
		}

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")

		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml)})
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
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		expectedOutput := `
# update: configmap/simple-cm (v1) namespace: kapp-test
---
apiVersion: v1
data:
  hello_msg: good-morning
kind: ConfigMap
metadata:
  creationTimestamp: "2022-05-11T15:27:14Z"
  labels:
    kapp.k14s.io/app: "1654768619446772000"
    kapp.k14s.io/association: v1.aa6ba70e7f14b3140de8009fda0a6fad
  name: simple-cm
  namespace: kapp-test
  resourceVersion: "201881"
  uid: ee57a857-1873-4c58-b021-8cc15bf4a3e6
# delete: configmap/simple-cm1 (v1) namespace: kapp-test

Succeeded
`

		// filtering value for label "kapp.k14s.io/app:" and replacing with hardcode value
		if len(strings.Split(out, "kapp.k14s.io/app: ")) > 1 {
			repValue := strings.Split((strings.Split(out, "kapp.k14s.io/app: ")[1]), "\n")[0]
			out = strings.Replace(out, repValue, `"1654768619446772000"`, 1)
		}

		// filtering value for label "creationTimestamp:" and replacing with hardcode value
		if len(strings.Split(out, "creationTimestamp: ")) > 1 {
			repValue := strings.Split((strings.Split(out, "creationTimestamp: ")[1]), "\n")[0]
			out = strings.Replace(out, repValue, `"2022-05-11T15:27:14Z"`, 1)
		}

		// filtering value for label "resourceVersion:" and replacing with hardcode value
		if len(strings.Split(out, "resourceVersion: ")) > 1 {
			repValue := strings.Split((strings.Split(out, "resourceVersion: ")[1]), "\n")[0]
			out = strings.Replace(out, repValue, `"201881"`, 1)
		}

		// filtering value for label "uid:" and replacing with hardcode value
		if len(strings.Split(out, "uid: ")) > 1 {
			repValue := strings.Split((strings.Split(out, "uid: ")[1]), "\n")[0]
			out = strings.Replace(out, repValue, "ee57a857-1873-4c58-b021-8cc15bf4a3e6", 1)
		}
		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))

		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")

	})

}
