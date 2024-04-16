// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"

	"carvel.dev/kapp/pkg/kapp/app"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestCreateUpdateDelete(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
`

	name := "test-create-update-delete"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy initial", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
	})

	logger.Section("deploy update with 1 delete, 1 update, 1 create", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)

		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))

		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)
	})

	logger.Section("delete application", func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{})

		NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config", env.Namespace, kubectl)
		NewMissingClusterResource(t, "configmap", "redis-config2", env.Namespace, kubectl)
	})
}

func TestCreateUpdateDelete_PrevApp(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
`

	yaml3 := `
---
apiVersion: v1
kind: Service
metadata:
  name: example-service
spec:
  ports:
  - port: 6381
    targetPort: 6381
  selector:
    app: example-app
    tier: backend
---
`

	appName := "test-create-update-delete-prev-app"
	prevAppName := "test-create-update-delete-prev-app-old"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", appName})
		kapp.Run([]string{"delete", "-a", prevAppName})
		os.Unsetenv("KAPP_FQ_CONFIGMAP_NAMES")
	}

	cleanUp()
	defer cleanUp()

	logger.Section("non-existent app, non-existent prevApp", func() {
		// creates app with name appName
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", appName, env.Namespace, kubectl)
		require.NotContains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		cleanUp()
	})

	logger.Section("existing unmigrated app", func() {
		logger.Section("deploy", func() {
			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

			NewMissingClusterResource(t, "service", "redis-primary", env.Namespace, kubectl)

			config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
			val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))

			require.Exactlyf(t, "value2", val, "Expected value to be updated")

			NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

			c := NewPresentClusterResource("configmap", appName, env.Namespace, kubectl)
			require.NotContains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

			NewMissingClusterResource(t, "configmap", prevAppName, env.Namespace, kubectl)
		})

		logger.Section("delete", func() {
			cleanUp()

			NewMissingClusterResource(t, "configmap", appName, env.Namespace, kubectl)
		})
	})

	logger.Section("existing migrated app", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		// Migrated app should get updated even if migration is disabled
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "False")
		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))
		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		cleanUp()
	})

	logger.Section("non-existent app, existing unmigrated prevApp", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))
		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", appName, env.Namespace, kubectl)
		require.NotContains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		NewMissingClusterResource(t, "configmap", prevAppName, env.Namespace, kubectl)

		cleanUp()
	})

	logger.Section("non-existent app, existing migrated prevApp", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		// Migrated prev-app should be renamed and updated
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "False")
		_, _ = kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName},
			RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", appName+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		NewMissingClusterResource(t, "configmap", prevAppName+app.AppSuffix, env.Namespace, kubectl)

		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		cleanUp()
	})

	logger.Section("delete migrated prevApp", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		kapp.Run([]string{"delete", "-a", "empty-app", "--prev-app", appName})

		NewMissingClusterResource(t, "configmap", appName, env.Namespace, kubectl)
		cleanUp()
	})

	logger.Section("delete unmigrated prevApp", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		kapp.Run([]string{"delete", "-a", "empty-app", "--prev-app", appName})

		NewMissingClusterResource(t, "configmap", appName, env.Namespace, kubectl)
		cleanUp()
	})

	// delete with --app and --prev-app existing should only delete --app and not --prev-app
	logger.Section("delete ignores prevApp", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml3)})

		kapp.Run([]string{"delete", "-a", appName, "--prev-app", prevAppName})

		NewMissingClusterResource(t, "configmap", appName, env.Namespace, kubectl)
		NewPresentClusterResource("configmap", prevAppName, env.Namespace, kubectl)
		cleanUp()
	})
}

func TestCreateUpdateDelete_PrevApp_FQConfigmap_Enabled(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: Service
metadata:
  name: redis-primary
spec:
  ports:
  - port: 6380
    targetPort: 6380
  selector:
    app: redis
    tier: backend
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value
`

	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  key: value2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config2
data:
  key: value
`

	appName := "test-create-update-delete-prev-app-fq-configmap"
	prevAppName := "test-create-update-delete-prev-app-fq-configmap-old"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", appName})
		kapp.Run([]string{"delete", "-a", prevAppName})
		os.Unsetenv("KAPP_FQ_CONFIGMAP_NAMES")
	}

	cleanUp()
	defer cleanUp()

	logger.Section("non-existent app, non-existent prevApp", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")

		// creates app with name appName
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		NewPresentClusterResource("service", "redis-primary", env.Namespace, kubectl)
		NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", appName+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		cleanUp()
	})

	logger.Section("existing unmigrated app", func() {
		logger.Section("deploy", func() {
			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

			os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
			kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

			config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
			val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))

			require.Exactlyf(t, "value2", val, "Expected value to be updated")

			NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

			c := NewPresentClusterResource("configmap", appName+app.AppSuffix, env.Namespace, kubectl)
			require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

			NewMissingClusterResource(t, "configmap", prevAppName, env.Namespace, kubectl)
		})

		logger.Section("delete", func() {
			cleanUp()

			NewMissingClusterResource(t, "configmap", appName+app.AppSuffix, env.Namespace, kubectl)
		})
	})

	logger.Section("existing migrated app", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		c := NewPresentClusterResource("configmap", appName+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		cleanUp()
	})

	logger.Section("non-existent app, existing unmigrated prevApp", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))
		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", appName+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		NewMissingClusterResource(t, "configmap", prevAppName, env.Namespace, kubectl)

		cleanUp()
	})

	logger.Section("non-existent app, existing migrated prevApp", func() {
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml2)})

		config := NewPresentClusterResource("configmap", "redis-config", env.Namespace, kubectl)
		val := config.RawPath(ctlres.NewPathFromStrings([]string{"data", "key"}))
		require.Exactlyf(t, "value2", val, "Expected value to be updated")

		NewPresentClusterResource("configmap", "redis-config2", env.Namespace, kubectl)

		c := NewPresentClusterResource("configmap", appName+app.AppSuffix, env.Namespace, kubectl)
		require.Contains(t, c.res.Annotations(), app.KappIsConfigmapMigratedAnnotationKey)

		NewMissingClusterResource(t, "configmap", prevAppName, env.Namespace, kubectl)

		cleanUp()
	})
}

func TestAppDeploy_With_Existing_OR_New_Res(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}
	kubectl := Kubectl{t, env.Namespace, logger}

	yaml1 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
data:
  key: value
`
	yaml2 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
data:
  key: value
`
	yaml3 := `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm2
data:
  key: value2
`

	name := "test-app-deploy-with-existing-res"
	name2 := "test-app-deploy-with-new-res"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", name})
		kapp.Run([]string{"delete", "-a", name2})
	}

	cleanUp()
	defer cleanUp()

	kubectl.RunWithOpts([]string{"apply", "-f", "-"},
		RunOpts{StdinReader: strings.NewReader(yaml1)})
	NewPresentClusterResource("configmap", "cm1", env.Namespace, kubectl)

	logger.Section("deploy app with existing resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--diff-changes"},
			RunOpts{StdinReader: strings.NewReader(yaml1)})

		expectedOutput := `
@@ update configmap/cm1 (v1) namespace: kapp-test @@
  ...
  4,  4   metadata:
  5     -   annotations:
  6     -     kubectl.kubernetes.io/last-applied-configuration: |
  7     -       {"apiVersion":"v1","data":{"key":"value"},"kind":"ConfigMap","metadata":{"annotations":{},"name":"cm1","namespace":"kapp-test"}}
  8,  5     creationTimestamp: "2006-01-02T15:04:05Z07:00"
      6 +   labels:
      7 +     kapp.k14s.io/app: "-replaced-"
      8 +     kapp.k14s.io/association: -replaced-
`
		out = replaceTimestampWithDfaultValue(out)
		out = replaceLabelValues(out)
		require.Contains(t, out, expectedOutput, "output does not match")
	})

	logger.Section("deploy app with all new resources", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2},
			RunOpts{StdinReader: strings.NewReader(yaml2)})
		NewPresentClusterResource("configmap", "cm2", env.Namespace, kubectl)
	})

	logger.Section("deploy again by updating resource", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name2, "--diff-changes"},
			RunOpts{StdinReader: strings.NewReader(yaml3)})

		expectedOutput := `
@@ update configmap/cm2 (v1) namespace: kapp-test @@
  ...
  1,  1   data:
  2     -   key: value
      2 +   key: value2
  3,  3   kind: ConfigMap
  4,  4   metadata:
`
		require.Contains(t, out, expectedOutput, "output does not match")
	})
}
