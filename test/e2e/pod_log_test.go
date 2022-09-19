// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPodLogs(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: simple-app
spec:
  selector:
    matchLabels:
      simple-app: ""
  serviceName: simple-app
  replicas: %d
  template:
    metadata:
      labels:
        simple-app: ""
      annotations:
        kapp.k14s.io/deploy-logs: %s
    spec:
      containers:
        - name: demo-container
          image: debian
          command: ["bash","-c","for i in {1..20}; do echo $i -n 'Carvel\n'; sleep 1;  done"]
`

	name := "test-pod-logs"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("Show logs for new Pods only when annotation value is default", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true,
			StdinReader: strings.NewReader(fmt.Sprintf(yaml, 1, ""))})
		NewPresentClusterResource("Pod", "simple-app-0", env.Namespace, kubectl)
		out, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, 2, ""))})
		NewPresentClusterResource("Pod", "simple-app-1", env.Namespace, kubectl)
		require.NotContains(t, out, "logs | simple-app-0 > demo-container | ", "Should not contain log for the existing Pod")
		require.Contains(t, out, "logs | simple-app-1 > demo-container | ", "Should contain log for the new Pod")
	})

	cleanUp()

	logger.Section("Show logs only for existing Pods with for-existing annotation value", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, 1, "for-existing"))})
		NewPresentClusterResource("Pod", "simple-app-0", env.Namespace, kubectl)
		out, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, 2, "for-existing"))})
		NewPresentClusterResource("Pod", "simple-app-1", env.Namespace, kubectl)
		require.Contains(t, out, "logs | simple-app-0 > demo-container | ", "Should contain log for the existing Pod")
		require.NotContains(t, out, "logs | simple-app-1 > demo-container | ", "Should not contain log for the new Pod")
	})
}
