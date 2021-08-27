// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
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
   namespace: __ns__ 
   annotations: 
     kapp.k14s.io/update-strategy: always-replace 
     kapp.k14s.io/change-group: job
 spec: 
   template: 
     metadata: 
       name: successful-job 
     spec: 
       containers: 
       - name: successful-job 
         image: busybox 
         command: ["/bin/sh", "-c", "sleep 5"] 
       restartPolicy: Never
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

	name := "test-wait-timeout"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Resource timed out waiting", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2m", "--wait-resource-timeout", "3s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if !strings.Contains(err.Error(), "Resource timed out waiting after 3s") {
			t.Fatalf("Expected to see timed out, but did not: '%s'", err.Error())
		}
	})

	cleanUp()

	logger.Section("Resource reconciled successfully before timeout", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2m", "--wait-resource-timeout", "15s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if err != nil {
			t.Fatalf("Expected to be successful without resource timeout: '%s'", err)
		}
	})

	cleanUp()

	logger.Section("Global timeout waiting", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"2s", "--wait-resource-timeout", "10s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if !strings.Contains(err.Error(), "kapp: Error: Timed out waiting after 2s") {
			t.Fatalf("Expected to see timed out, but did not: '%s'", err.Error())
		}
	})

	cleanUp()

	logger.Section("Resource reconciled successfully before timeout", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"1m", "--wait-resource-timeout", "15s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if err != nil {
			t.Fatalf("Expected to be successful without global timeout: '%s'", err)
		}
	})
}
