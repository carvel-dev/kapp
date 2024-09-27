// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func TestApplyWaitErrors(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job-fail
  annotations:
    kapp.k14s.io/change-group: "job-fail"
spec:
  template:
    metadata:
      name: my-job-fail
    spec:
      restartPolicy: Never
      containers:
        - name: my-job-fail
          image: busybox
          command: [ "sh", "-c", "exit 1" ]
  backoffLimit: 0
---
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job-fail-2
  annotations:
    kapp.k14s.io/change-group: "job-fail"
spec:
  template:
    metadata:
      name: my-job-fail-2
    spec:
      restartPolicy: Never
      containers:
        - name: my-job-fail-2
          image: busybox
          command: [ "sh", "-c", "exit 1" ]
  backoffLimit: 0
---
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job-succeed
  annotations:
    kapp.k14s.io/change-group: "job-succeed"
    kapp.k14s.io/change-rule: "upsert before upserting service-not-deployed"
spec:
  template:
    metadata:
      name: my-job-succeed
    spec:
      restartPolicy: Never
      containers:
        - name: my-job-succeed
          image: busybox
          command: [ "sh", "-c", "exit 0" ]
  backoffLimit: 0
---
apiVersion: v1
kind: Service
metadata:
  name: service-succeed
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting job-succeed"
spec:
  ports:
  - port: 80
---
apiVersion: v1
kind: Service
metadata:
  name: service-fail
  annotations:
    kapp.k14s.io/change-group: "service-fail"
    kapp.k14s.io/change-rule: "upsert after upserting job-succeed"
---
apiVersion: v1
kind: Service
metadata:
  name: service-not-deployed
  annotations:
    kapp.k14s.io/change-group: "service-not-deployed"
    kapp.k14s.io/change-rule: "upsert after upserting job-fail"
spec:
  ports:
  - port: 80
---
apiVersion: v1
kind: Service
metadata:
  name: service-not-deployed-2
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting service-fail"
spec:
  ports:
  - port: 80
`

	name := "test-multiple-errors"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy with multiple errors and exit early on error set to false", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--exit-early-on-apply-error=false", "--exit-early-on-wait-error=false"},
			RunOpts{StdinReader: strings.NewReader(yaml1), AllowError: true})

		expectedErrs := []string{
			`kapp: Error:`,
			fmt.Sprintf(`- waiting on reconcile job/my-job-fail-2 (batch/v1) namespace: %s: Finished waiting unsuccessfully: Failed with reason BackoffLimitExceeded: Job has reached the specified backoff limit`, env.Namespace),
			fmt.Sprintf(`- waiting on reconcile job/my-job-fail (batch/v1) namespace: %s: Finished waiting unsuccessfully: Failed with reason BackoffLimitExceeded: Job has reached the specified backoff limit`, env.Namespace),
			fmt.Sprintf(`- create service/service-fail (v1) namespace: %s: Creating resource service/service-fail (v1) namespace: kapp-test: API server says: Service "service-fail" is invalid: spec.ports: Required value (reason: Invalid)`, env.Namespace),
		}

		for _, expectedErr := range expectedErrs {
			require.Containsf(t, err.Error(), expectedErr, "Expected to see expected err in output, but did not")
		}

		out := kapp.Run([]string{"inspect", "-a", name, "--filter-kind", "Service", "--json"})

		expectedResources := []map[string]string{{
			"age":             "<replaced>",
			"kind":            "Service",
			"name":            "service-succeed",
			"namespace":       env.Namespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		inspectOut := uitest.JSONUIFromBytes(t, []byte(out))

		require.Exactlyf(t, expectedResources, replaceAge(inspectOut.Tables[0].Rows), "Expected to see correct changes")
	})
}

func TestExitEarlyOnApplyErrorFlag(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: service-fail
---
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job-succeed
  annotations:
    kapp.k14s.io/change-group: "job-succeed"
spec:
  template:
    metadata:
      name: my-job-succeed
    spec:
      restartPolicy: Never
      containers:
        - name: my-job-succeed
          image: busybox
          command: [ "sh", "-c", "exit 0" ]
  backoffLimit: 0
---
apiVersion: v1
kind: Service
metadata:
  name: service-succeed
  annotations:
    kapp.k14s.io/change-rule: "upsert after upserting job-succeed"
spec:
  ports:
  - port: 80
`

	name := "test-exit-early-on-apply-error"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	expectedErr := strings.ReplaceAll(`kapp: Error: create service/service-fail (v1) namespace: <test-ns>:
  Creating resource service/service-fail (v1) namespace: <test-ns>:
    API server says:
      Service "service-fail" is invalid: spec.ports:
        Required value (reason: Invalid)
`, "<test-ns>", env.Namespace)

	logger.Section("deploy with exit-early-on-apply-error=false", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--exit-early-on-apply-error=false"},
			RunOpts{StdinReader: strings.NewReader(yaml1), AllowError: true})

		require.Containsf(t, err.Error(), expectedErr, "Expected to see expected err in output, but did not")

		out := kapp.Run([]string{"inspect", "-a", name, "--json", "--filter-kind", "Service"})

		expectedResources := []map[string]string{{
			"age":             "<replaced>",
			"kind":            "Service",
			"name":            "service-succeed",
			"namespace":       env.Namespace,
			"owner":           "kapp",
			"reconcile_info":  "",
			"reconcile_state": "ok",
		}}

		inspectOut := uitest.JSONUIFromBytes(t, []byte(out))

		require.Exactlyf(t, expectedResources, replaceAge(inspectOut.Tables[0].Rows), "Expected to see correct changes")
	})

	cleanUp()

	logger.Section("deploy with exit-early-on-apply-error=true", func() {
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml1), AllowError: true})

		require.Containsf(t, err.Error(), expectedErr, "Expected to see expected err in output, but did not")

		out := kapp.Run([]string{"inspect", "-a", name, "--json", "--filter-kind", "Service"})

		expectedResources := []map[string]string{}

		inspectOut := uitest.JSONUIFromBytes(t, []byte(out))

		require.Exactlyf(t, expectedResources, replaceAge(inspectOut.Tables[0].Rows), "Expected to see correct changes")
	})
}
