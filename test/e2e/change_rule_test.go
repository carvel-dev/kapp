// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangeRuleWaitsForUnchangedJob(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	name := "test-change-rule-unchanged-job"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}
	cleanUp()
	defer cleanUp()

	// Suspending the Job keeps it pending without relying on pod startup timing.
	prerequisite := `
apiVersion: batch/v1
kind: Job
metadata:
  name: change-rule-prerequisite
  annotations:
    kapp.k14s.io/change-group: prerequisite
spec:
  suspend: true
  template:
    spec:
      containers:
      - name: job
        image: busybox
        command: ["true"]
      restartPolicy: Never
`
	config := prerequisite + `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: change-rule-dependent
  annotations:
    kapp.k14s.io/change-rule: upsert after upserting prerequisite
`

	// Simulate an interrupted deployment that created only the prerequisite.
	_, err := kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-", "--wait=false"},
		RunOpts{IntoNs: true, StdinReader: strings.NewReader(prerequisite)})
	require.NoError(t, err)

	out, err := kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-", "--wait-timeout=1s"},
		RunOpts{IntoNs: true, StdinReader: strings.NewReader(config), AllowError: true})
	require.ErrorContains(t, err, "Timed out waiting after 1s")
	require.Contains(t, out, "noop job/change-rule-prerequisite")
	exists, err := ClusterResourceExists("configmap", "change-rule-dependent", kubectl)
	require.NoError(t, err)
	require.False(t, exists, "the dependent must not be created while the unchanged Job is pending")

	// Explicitly disabling waiting must still allow the dependent to be created.
	_, err = kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-", "--wait=false"},
		RunOpts{IntoNs: true, StdinReader: strings.NewReader(config)})
	require.NoError(t, err)
	exists, err = ClusterResourceExists("configmap", "change-rule-dependent", kubectl)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestServiceAccountAssociatedSecretDefaultChangeRule(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	config := `
apiVersion: v1
kind: Namespace
metadata:
  name: kapp-token-test
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-0
  namespace: kapp-token-test
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-0
  namespace: kapp-token-test
  annotations:
    kubernetes.io/service-account.name: sa-0
type: kubernetes.io/service-account-token
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-1
  namespace: kapp-token-test
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-1
  namespace: kapp-token-test
  annotations:
    kubernetes.io/service-account.name: sa-1
type: kubernetes.io/service-account-token
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-2
  namespace: kapp-token-test
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-2
  namespace: kapp-token-test
  annotations:
    kubernetes.io/service-account.name: sa-2
type: kubernetes.io/service-account-token
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-3
  namespace: kapp-token-test
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-3
  namespace: kapp-token-test
  annotations:
    kubernetes.io/service-account.name: sa-3
type: kubernetes.io/service-account-token
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-4
  namespace: kapp-token-test
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-4
  namespace: kapp-token-test
  annotations:
    kubernetes.io/service-account.name: sa-4
type: kubernetes.io/service-account-token
`

	name := "test-sa-created-before-secret-associated-with-sa"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()

	logger.Section("deploy resource with secret associated with service account", func() {
		// deploying it multiple times to make sure it's able to find secret everytime
		for i := 0; i < 5; i++ {
			_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{
				StdinReader: strings.NewReader(config),
			})
			if err != nil {
				require.Errorf(t, err, "Expected resources to be created")
			}
			cleanUp()
		}
	})
}
