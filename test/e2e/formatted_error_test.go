// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormattedError(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: batch/v1
kind: Job
metadata:
  name: successful-job
  namespace: default
spec:
  selector:
    matchLabels:
      blah: balh
  template:
    metadata:
      name: successful-job
      labels:
        foo: foo
    spec:
      containers:
      - name: successful-job
        image: busybox
        command: ["/bin/sh", "-c", "sleep 11"]
      restartPolicy: Never
`

	name := "test-formatted-error"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy with errors", func() {
		expectedErr := strings.TrimSpace(`
kapp: Error: Applying create job/successful-job (batch/v1) namespace: default:
  Creating resource job/successful-job (batch/v1) namespace: default:
    Job.batch "successful-job" is invalid: 

  - spec.selector: Invalid value: v1.LabelSelector{MatchLabels:map[string]string{"blah":"balh", -replaced-, -replaced-}, MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: 'selector' not auto-generated

  - spec.template.metadata.labels: Invalid value: map[string]string{-replaced-, "foo":"foo", "job-name":"successful-job", -replaced-, -replaced-}: 'selector' does not match template 'labels'

 (reason: Invalid)
`)

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml1), AllowError: true})

		out := strings.ReplaceAll(err.Error(), "`", "'")

		replaceAnns := regexp.MustCompile(`"kapp\.k14s\.io\/(app|association)":"[^"]+"`)
		out = replaceAnns.ReplaceAllString(out, "-replaced-")

		replaceUIDs := regexp.MustCompile(`"controller-uid":"[^"]+"`)
		out = replaceUIDs.ReplaceAllString(out, "-replaced-")

		require.Containsf(t, out, expectedErr, "Expected to see expected err in output, but did not")
	})
}
