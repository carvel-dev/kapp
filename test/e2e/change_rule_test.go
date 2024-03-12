// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
