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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-app2
spec:
  replicas: %d
  selector:
    matchLabels:
      simple-app2: ""
  template:
    metadata:
      labels:
        simple-app2: ""
      annotations:
        kapp.k14s.io/deploy-logs: %s
    spec:
      containers:
        - name: demo-container
          image: debian
          command: ["bash","-c","for i in {1..10}; do echo $i -n 'Carvel\n'; sleep 1;  done"]
`

	name := "test-pod-logs"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	var existingPodLog string

	logger.Section("Should only show log for new Pod with replicas 1", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true,
			StdinReader: strings.NewReader(fmt.Sprintf(yaml, 1, ""))})
		resources, _ := GetClusterResourcesByKind("Pod", "default", kubectl)
		if len(resources) > 0 {
			existingPodLog = fmt.Sprintf("logs | %s > demo-container | ", resources[0].Name())
			require.Contains(t, out, existingPodLog, "Expected to show log for the first Pod")
		}
	})

	logger.Section("Should not show log for existing Pod when replicas changed to 2", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(fmt.Sprintf(yaml, 2, "for-new"))})
		if existingPodLog != "" {
			require.NotContains(t, out, existingPodLog, "Should not contain log for the existing Pod")
		}
	})
}
