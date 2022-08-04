// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPriodicUpdate(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}
	//fieldsExcludedInMatch := []string{"kapp.k14s.io/app", "creationTimestamp:", "resourceVersion:", "uid:", "selfLink:", "kapp.k14s.io/association"}

	yaml := `
--- 
apiVersion: v1
data:
  hello_msg: carvel
kind: ConfigMap
metadata:
  annotations:
    kapp.k14s.io/max-duration: 4s
  name: simple-cm1
---
apiVersion: v1
data: 
  hello_msg: carvel
kind: ConfigMap
metadata: 
  annotations: 
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/max-duration: 4s
  name: simple-cm2
`

	name := "test-priodic-update"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Initial deploy", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{StdinReader: strings.NewReader(yaml)})
		NewPresentClusterResource("ConfigMap", "simple-cm1", env.Namespace, kubectl)
		NewPresentClusterResource("ConfigMap", "simple-cm2-ver-1", env.Namespace, kubectl)

	})

	time.Sleep(1 * time.Second)
	logger.Section("Deploy again after 3 second deploy", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--diff-run", "--diff-summary=false", "--tty=false"}, RunOpts{StdinReader: strings.NewReader(yaml)})
		require.Equal(t, "", out, "output does not match")
	})

	time.Sleep(3 * time.Second)
	logger.Section("Deploy again after 3 second deploy", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes", "--diff-run", "--diff-summary=false"}, RunOpts{StdinReader: strings.NewReader(yaml)})

		expectedOutput := `
@@ create configmap/simple-cm2-ver-2 (v1) namespace: kapp-test @@
  ...
  5,  5     annotations:
      6 +     kapp.k14s.io/last-renewed-time: "2006-01-02T15:04:05Z07:00"
  6,  7       kapp.k14s.io/max-duration: 4s
  7,  8       kapp.k14s.io/versioned: ""
@@ update configmap/simple-cm1 (v1) namespace: kapp-test @@
  ...
  5,  5     annotations:
      6 +     kapp.k14s.io/last-renewed-time: "2006-01-02T15:04:05Z07:00"
  6,  7       kapp.k14s.io/max-duration: 4s
  7,  8     creationTimestamp: "2006-01-02T15:04:05Z07:00"
`

		out = strings.TrimSpace(replaceTarget(replaceSpaces(replaceTs(out))))
		out = replaceString(out)
		expectedOutput = strings.TrimSpace(replaceSpaces(expectedOutput))
		require.Contains(t, out, expectedOutput, "output does not match")
	})

}

// clearKeys will remove all matching fields in fieldsExcludedInMatch from out
func replaceString(out string) string {
	r := regexp.MustCompile("[1-9]\\d{3}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z")
	out = r.ReplaceAllString(out, time.RFC3339)
	return out
}
