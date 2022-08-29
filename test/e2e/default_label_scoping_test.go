// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultLabelScopingRulesFlag(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, Logger{}}

	name := "test-default-label-scoping-rules"

	yaml := `
---
apiVersion: v1
kind: Service
metadata:
  name: simple-app
spec:
  ports:
  - port: 80
    targetPort: 80
  selector:
    simple-app: ""
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-app
spec:
  selector:
    matchLabels:
      simple-app: ""
  template:
    metadata:
      labels:
        simple-app: ""
    spec:
      containers:
      - name: simple-app
        image: docker.io/dkalinin/k8s-simple-app@sha256:4c8b96d4fffdfae29258d94a22ae4ad1fe36139d47288b8960d9958d1e63a9d0
        env:
        - name: HELLO_MSG
          value: stranger
`
	config := `
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

labelScopingRules:
- path: [spec, selector]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}
`

	logger.Section("deploying app with default-label-scoping-rules", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--diff-run"},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOut := `
@@ create service/simple-app (v1) namespace: <test-ns> @@
      0 + apiVersion: v1
      1 + kind: Service
      2 + metadata:
      3 +   labels:
      4 +     kapp.k14s.io/app: "-replaced-"
      5 +     kapp.k14s.io/association: -replaced-
      6 +   name: simple-app
      7 +   namespace: <test-ns>
      8 + spec:
      9 +   ports:
     10 +   - port: 80
     11 +     targetPort: 80
     12 +   selector:
     13 +     kapp.k14s.io/app: "-replaced-"
     14 +     simple-app: ""
     15 + 
@@ create deployment/simple-app (apps/v1) namespace: <test-ns> @@
      0 + apiVersion: apps/v1
      1 + kind: Deployment
      2 + metadata:
      3 +   labels:
      4 +     kapp.k14s.io/app: "-replaced-"
      5 +     kapp.k14s.io/association: -replaced-
      6 +   name: simple-app
      7 +   namespace: <test-ns>
      8 + spec:
      9 +   selector:
     10 +     matchLabels:
     11 +       kapp.k14s.io/app: "-replaced-"
     12 +       simple-app: ""
     13 +   template:
     14 +     metadata:
     15 +       labels:
     16 +         kapp.k14s.io/app: "-replaced-"
     17 +         kapp.k14s.io/association: -replaced-
     18 +         simple-app: ""
`
		expectedOut = strings.ReplaceAll(expectedOut, "<test-ns>", env.Namespace)
		require.Contains(t, replaceLabelValues(out), expectedOut, "Expected labels to match")
	})

	logger.Section("deploying app with default-label-scoping-rules=false", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--diff-run", "--default-label-scoping-rules=false"},
			RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOut := `
@@ create service/simple-app (v1) namespace: <test-ns> @@
      0 + apiVersion: v1
      1 + kind: Service
      2 + metadata:
      3 +   labels:
      4 +     kapp.k14s.io/app: "-replaced-"
      5 +     kapp.k14s.io/association: -replaced-
      6 +   name: simple-app
      7 +   namespace: <test-ns>
      8 + spec:
      9 +   ports:
     10 +   - port: 80
     11 +     targetPort: 80
     12 +   selector:
     13 +     simple-app: ""
     14 + 
@@ create deployment/simple-app (apps/v1) namespace: <test-ns> @@
      0 + apiVersion: apps/v1
      1 + kind: Deployment
      2 + metadata:
      3 +   labels:
      4 +     kapp.k14s.io/app: "-replaced-"
      5 +     kapp.k14s.io/association: -replaced-
      6 +   name: simple-app
      7 +   namespace: <test-ns>
      8 + spec:
      9 +   selector:
     10 +     matchLabels:
     11 +       simple-app: ""
     12 +   template:
     13 +     metadata:
     14 +       labels:
     15 +         kapp.k14s.io/app: "-replaced-"
     16 +         kapp.k14s.io/association: -replaced-
     17 +         simple-app: ""
`
		expectedOut = strings.ReplaceAll(expectedOut, "<test-ns>", env.Namespace)
		require.Contains(t, replaceLabelValues(out), expectedOut, "Expected labels to match")
	})

	logger.Section("deploying app with default-label-scoping-rules=false and custom label scoping rules", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "-c", "--diff-run", "--default-label-scoping-rules=false"},
			RunOpts{StdinReader: strings.NewReader(yaml + config)})

		expectedOut := `
@@ create service/simple-app (v1) namespace: <test-ns> @@
      0 + apiVersion: v1
      1 + kind: Service
      2 + metadata:
      3 +   labels:
      4 +     kapp.k14s.io/app: "-replaced-"
      5 +     kapp.k14s.io/association: -replaced-
      6 +   name: simple-app
      7 +   namespace: <test-ns>
      8 + spec:
      9 +   ports:
     10 +   - port: 80
     11 +     targetPort: 80
     12 +   selector:
     13 +     kapp.k14s.io/app: "-replaced-"
     14 +     simple-app: ""
     15 + 
@@ create deployment/simple-app (apps/v1) namespace: <test-ns> @@
      0 + apiVersion: apps/v1
      1 + kind: Deployment
      2 + metadata:
      3 +   labels:
      4 +     kapp.k14s.io/app: "-replaced-"
      5 +     kapp.k14s.io/association: -replaced-
      6 +   name: simple-app
      7 +   namespace: <test-ns>
      8 + spec:
      9 +   selector:
     10 +     matchLabels:
     11 +       simple-app: ""
     12 +   template:
     13 +     metadata:
     14 +       labels:
     15 +         kapp.k14s.io/app: "-replaced-"
     16 +         kapp.k14s.io/association: -replaced-
     17 +         simple-app: ""
`
		expectedOut = strings.ReplaceAll(expectedOut, "<test-ns>", env.Namespace)
		require.Contains(t, replaceLabelValues(out), expectedOut, "Expected labels to match")
	})
}

func replaceLabelValues(in string) string {
	replaceAnns := regexp.MustCompile("kapp\\.k14s\\.io\\/app: .+")
	in = replaceAnns.ReplaceAllString(in, `kapp.k14s.io/app: "-replaced-"`)
	replaceAnns = regexp.MustCompile("kapp\\.k14s\\.io\\/association: .+")
	return replaceAnns.ReplaceAllString(in, `kapp.k14s.io/association: -replaced-`)
}
