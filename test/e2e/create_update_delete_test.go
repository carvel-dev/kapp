// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/k14s/kapp/pkg/kapp/app"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", prevAppName}, RunOpts{IntoNs: true, StdinReader: strings.NewReader(yaml1)})

		// KAPP_FQ_CONFIGMAP_NAMES=False is not supported - must go from migrated => migrated
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "False")
		out := kapp.Run([]string{"ls", "--json"})

		expectedAppsList := []map[string]string{{
			"last_change_age":        "<replaced>",
			"last_change_successful": "true",
			"name":                   prevAppName + app.AppSuffix,
			"namespaces":             env.Namespace,
		}}

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		require.Equalf(t, expectedAppsList, replaceLastChangeAge(resp.Tables[0].Rows), "Expected to match")

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml2)})

		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "is already associated with a different app", "Expected app to be associated with another owner")

		logger.Section("delete", func() {
			os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
			cleanUp()

			NewMissingClusterResource(t, "configmap", prevAppName, env.Namespace, kubectl)
		})
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

		// KAPP_FQ_CONFIGMAP_NAMES=False is not supported - must go from migrated => migrated
		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "False")
		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", appName, "--prev-app", prevAppName},
			RunOpts{IntoNs: true, AllowError: true, StdinReader: strings.NewReader(yaml2)})

		require.Errorf(t, err, "Expected to receive error")
		require.Containsf(t, err.Error(), "is already associated with a different app", "Expected app to be associated with another owner")

		os.Setenv("KAPP_FQ_CONFIGMAP_NAMES", "True")
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
