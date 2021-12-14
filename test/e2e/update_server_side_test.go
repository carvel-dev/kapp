// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSAUpdateSimple(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    name: p0
  selector:
    app: redis
`

	kubectlChange := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6381
    name: p1
  selector:
    app: redis
`

	yaml2 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    name: p0
  - port: 6382
    name: p2
  selector:
    app: redis
`

	yamlExpected, _ := yaml.YAMLToJSON([]byte(`
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    name: p0
  - port: 6381
    name: p1
  - port: 6382
    name: p2
  selector:
    app: redis
`))

	name := strings.ToLower(t.Name())
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	kubectl := Kubectl{t, env.Namespace, logger}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy basic service", func() {
		_, err := kapp.RunEmbedded([]string{"deploy", "--server-side", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1), StdoutWriter: os.Stdout})
		require.NoError(t, err)
	})

	logger.Section("edit resource with kubectl outside of kapp", func() {
		_, err := kubectl.RunWithOpts([]string{"apply", "--validate=false", "--server-side", "-f", "-"},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(kubectlChange)},
		)
		require.NoError(t, err)
	})

	logger.Section("deploy updated service", func() {
		_, err := kapp.RunEmbedded([]string{"deploy", "--server-side", "-f", "-", "-a", name},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2), StdoutWriter: os.Stdout})
		require.NoError(t, err)
	})

	inClusterObj := corev1.Service{}

	err := kubectl.RunWithOptsIntoJSON([]string{"get", "svc", "redis-primary"},
		RunOpts{IntoNs: true}, &inClusterObj)
	require.NoError(t, err)

	tmpFile := newTmpFile(string(yamlExpected), t)
	defer os.Remove(tmpFile.Name())

	expectedObj := corev1.Service{}

	// Patch dry run returns merged object with all patch-file fields present
	err = kubectl.RunWithOptsIntoJSON([]string{"patch", "svc", "redis-primary", "--patch-file", tmpFile.Name(), "--dry-run=client"},
		RunOpts{IntoNs: true}, &expectedObj)
	require.NoError(t, err)

	require.Equal(t, expectedObj, inClusterObj)

}
