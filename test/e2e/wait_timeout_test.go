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

		if !strings.Contains(err.Error(), "Resource timed out waiting after 1s") {
			t.Fatalf("Expected to see timed out, but did not: '%s'", err.Error())
		}
	})

	cleanUp()

	logger.Section("Global timeout waiting", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"1s", "--wait-resource-timeout", "100s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if !strings.Contains(err.Error(), "kapp: Error: Timed out waiting after 1s") {
			t.Fatalf("Expected to see timed out, but did not: '%s'", err.Error())
		}
	})

	cleanUp()

	logger.Section("Resource reconciled successfully before timeout", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--wait-timeout",
			"10000s", "--wait-resource-timeout", "10000s", "--json"},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml1)})

		if err != nil {
			t.Fatalf("Expected to be successful without resource timeout: '%s'", err)
		}
	})
}
