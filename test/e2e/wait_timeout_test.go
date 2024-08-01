// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWaitTimeout(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
 apiVersion: batch/v1 
 kind: Job 
 metadata: 
   name: successful-job
 spec: 
   template: 
     metadata: 
       name: successful-job 
     spec: 
       containers: 
       - name: successful-job 
         image: busybox
         command: ["/bin/sh", "-c", "sleep 10"] 
       restartPolicy: Never
`

	yaml2 := `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: %s
    ports:
    - containerPort: 80
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
waitRules:
- supportsObservedGeneration: true
  conditionMatchers:
  - type: ContainersReady
    status: "False"
    timeout: 50s
  - type: Ready
    status: "True"
    success: true
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}
`

	name := "test-wait-timeout"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Resource timed out waiting", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"100s", "--wait-resource-timeout", "1s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.Containsf(t, err.Error(), "Resource timed out waiting after 1s", "Expected to see timed out, but did not")
	})

	cleanUp()

	logger.Section("Global timeout waiting", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"1s", "--wait-resource-timeout", "100s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.Containsf(t, err.Error(), "kapp: Error: Timed out waiting after 1s", "Expected to see timed out, but did not")
	})

	cleanUp()

	logger.Section("Resource reconciled successfully before timeout", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"10000s", "--wait-resource-timeout", "10000s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		require.NoErrorf(t, err, "Expected to be successful without resource timeout")
	})

	cleanUp()

	logger.Section("Deploy timeout after staying in a condition for certain time", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml2, "nginx:200"))})

		require.Error(t, err)
		require.Contains(t, err.Error(), "message: containers with unready status: [nginx]) continuously for 50s duration")
	})

	cleanUp()

	logger.Section("Deploy should be successful", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml2, "nginx"))})

		require.NoError(t, err)
	})
}
