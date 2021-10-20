// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	depYAML := `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  selector:
    matchLabels:
      app: dep
  replicas: 1
  template:
    metadata:
      labels:
        app: dep
    spec:
      containers:
      - name: echo
        image: hashicorp/http-echo
        args:
        - -listen=:80
        - -text=hello
        ports:
        - containerPort: 80
        envFrom:
        - configMapRef:
            name: config
      initContainers:
      - name: echo-init
        image: hashicorp/http-echo
        args:
        - -version
        envFrom:
        - configMapRef:
            name: config
      volumes:
      - name: vol1
        secret:
          secretName: secret
`

	yaml1 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val1
` + depYAML

	yaml2 := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val2
---
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
data:
  key1: val2
` + depYAML

	expectedYAML1Diff := `
@@ create configmap/config-ver-1 (v1) namespace: kapp-test @@
-linesss- apiVersion: v1
-linesss- data:
-linesss-   key1: val1
-linesss- kind: ConfigMap
-linesss- metadata:
-linesss-   annotations:
-linesss-     kapp.k14s.io/versioned: ""
-linesss-   labels:
-linesss-     -replaced-
-linesss-     -replaced-
-linesss-   name: config-ver-1
-linesss-   namespace: kapp-test
-linesss- 
@@ create secret/secret-ver-1 (v1) namespace: kapp-test @@
-linesss- apiVersion: v1
-linesss- data:
-linesss-   key1: val1
-linesss- kind: Secret
-linesss- metadata:
-linesss-   annotations:
-linesss-     kapp.k14s.io/versioned: ""
-linesss-   labels:
-linesss-     -replaced-
-linesss-     -replaced-
-linesss-   name: secret-ver-1
-linesss-   namespace: kapp-test
-linesss- 
@@ create deployment/dep (apps/v1) namespace: kapp-test @@
-linesss- apiVersion: apps/v1
-linesss- kind: Deployment
-linesss- metadata:
-linesss-   labels:
-linesss-     -replaced-
-linesss-     -replaced-
-linesss-   name: dep
-linesss-   namespace: kapp-test
-linesss- spec:
-linesss-   replicas: 1
-linesss-   selector:
-linesss-     matchLabels:
-linesss-       app: dep
-linesss-       -replaced-
-linesss-   template:
-linesss-     metadata:
-linesss-       labels:
-linesss-         app: dep
-linesss-         -replaced-
-linesss-         -replaced-
-linesss-     spec:
-linesss-       containers:
-linesss-       - args:
-linesss-         - -listen=:80
-linesss-         - -text=hello
-linesss-         envFrom:
-linesss-         - configMapRef:
-linesss-             name: config-ver-1
-linesss-         image: hashicorp/http-echo
-linesss-         name: echo
-linesss-         ports:
-linesss-         - containerPort: 80
-linesss-       initContainers:
-linesss-       - args:
-linesss-         - -version
-linesss-         envFrom:
-linesss-         - configMapRef:
-linesss-             name: config-ver-1
-linesss-         image: hashicorp/http-echo
-linesss-         name: echo-init
-linesss-       volumes:
-linesss-       - name: vol1
-linesss-         secret:
-linesss-           secretName: secret-ver-1
-linesss- 
`

	expectedYAML2Diff := `
@@ create configmap/config-ver-2 (v1) namespace: kapp-test @@
  ...
-linesss- data:
-linesss-   key1: val1
-linesss-   key1: val2
-linesss- kind: ConfigMap
-linesss- metadata:
@@ create secret/secret-ver-2 (v1) namespace: kapp-test @@
  ...
-linesss- data:
-linesss-   key1: val1
-linesss-   key1: val2
-linesss- kind: Secret
-linesss- metadata:
@@ update deployment/dep (apps/v1) namespace: kapp-test @@
  ...
-linesss-         - configMapRef:
-linesss-             name: config-ver-1
-linesss-             name: config-ver-2
-linesss-         image: hashicorp/http-echo
-linesss-         name: echo
  ...
-linesss-         - configMapRef:
-linesss-             name: config-ver-1
-linesss-             name: config-ver-2
-linesss-         image: hashicorp/http-echo
-linesss-         name: echo-init
  ...
-linesss-         secret:
-linesss-           secretName: secret-ver-1
-linesss-           secretName: secret-ver-2
-linesss- status:
-linesss-   availableReplicas: 1
`

	name := "test-template"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	depPath := []interface{}{"spec", "template", "spec", "containers", 0, "envFrom", 0, "configMapRef", "name"}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--tty", "--diff-mask=false"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		checkChangesOutput(t, out, expectedYAML1Diff)

		dep := NewPresentClusterResource("deployment", "dep", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-1", env.Namespace, kubectl)

		val := dep.RawPath(ctlres.NewPathFromInterfaces(depPath))

		require.Exactlyf(t, "config-ver-1", val, "Expected value to be updated")
	})

	logger.Section("deploy update that changes configmap", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--tty", "--diff-mask=false"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		checkChangesOutput(t, out, expectedYAML2Diff)

		dep := NewPresentClusterResource("deployment", "dep", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-2", env.Namespace, kubectl)

		val := dep.RawPath(ctlres.NewPathFromInterfaces(depPath))

		require.Exactlyf(t, "config-ver-2", val, "Expected value to be updated")
	})

	logger.Section("deploy update that has no changes", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--tty", "--diff-mask=false"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})
		checkChangesOutput(t, out, "")

		dep := NewPresentClusterResource("deployment", "dep", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-1", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "config-ver-2", env.Namespace, kubectl)

		val := dep.RawPath(ctlres.NewPathFromInterfaces(depPath))

		require.Exactlyf(t, "config-ver-2", val, "Expected value to be updated")
	})

	// TODO deploy via patch or filter
}

func checkChangesOutput(t *testing.T, actualOutput, expectedOutput string) {
	replaceAnns := regexp.MustCompile("kapp\\.k14s\\.io\\/(app|association): .+")
	actualOutput = replaceAnns.ReplaceAllString(actualOutput, "-replaced-")

	actualOutput = strings.TrimSpace(strings.Split(replaceTarget(actualOutput), "Changes")[0])
	expectedOutput = strings.TrimSpace(expectedOutput)

	// Line numbers may change depending on what's being added to metadata section for example
	// (metadata.managedFields was added and threw off all lines numbers)
	diffLinesRegexp := regexp.MustCompile(`(?m:^\s*(\d{1,3}\s*|\d{1,3},\s*\d{1,3}|\d{1,3}) [\-+ ])`)
	actualOutput = diffLinesRegexp.ReplaceAllString(actualOutput, "-linesss-")

	// Useful for debugging:
	// printLines("actual", actualOutput)
	// printLines("expected", expectedOutput)

	require.Equalf(t, expectedOutput, actualOutput, "Expected output to match actual")
}
