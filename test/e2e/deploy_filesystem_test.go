// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"carvel.dev/kapp/pkg/kapp/cmd/app"
	"carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"carvel.dev/kapp/pkg/kapp/preflight"
)

func TestDeployFilesystem(t *testing.T) {
	env := BuildEnv(t)
	testLogger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, testLogger}
	appName := "test-deploy-filesystem"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", appName})
	}

	cleanUp()
	t.Cleanup(cleanUp)

	theUI := ui.NewConfUI(ui.NewNoopLogger())
	theUI.EnableNonInteractive()

	configFactory := core.NewConfigFactoryImpl()
	configFactory.ConfigureClient(100, 200)

	// We don't need to customize any of these, but we'll get panics if they're nil
	configFactory.ConfigurePathResolver(func() (string, error) {
		return "", nil
	})
	configFactory.ConfigureContextResolver(func() (string, error) {
		return "", nil
	})
	configFactory.ConfigureYAMLResolver(func() (string, error) {
		return "", nil
	})

	depsFactory := core.NewDepsFactoryImpl(configFactory, theUI)
	log := logger.NewUILogger(theUI)

	deployOptions := app.NewDeployOptions(theUI, depsFactory, log, &preflight.Registry{})
	deployOptions.AppFlags.NamespaceFlags.Name = env.Namespace
	deployOptions.AppFlags.Name = appName
	deployOptions.FileFlags.Files = []string{
		"dir1",            // directory
		"file1.yaml",      // file in root of fs
		"dir2/file2.yaml", // file in subdir
		"dir3",            // another directory
	}

	flagsFactory := core.NewFlagsFactory(configFactory, depsFactory)

	deployCmd := app.NewDeployCmd(deployOptions, flagsFactory)

	now := time.Now().Unix()
	labelSelector := fmt.Sprintf("now=%d", now)

	inMemFS := newFSBuilder().
		file(
			"dir1/cm1.yaml",
			testConfigMap(env.Namespace, "cm1", now),
		).
		file(
			"dir1/cm2.yaml",
			testConfigMap(env.Namespace, "cm2", now),
		).
		file(
			"file1.yaml",
			testConfigMap(env.Namespace, "cm3", now),
		).
		file(
			"dir2/file2.yaml",
			testConfigMap(env.Namespace, "cm4", now),
		).
		file(
			"dir3/cm5.yaml",
			testConfigMap(env.Namespace, "cm5", now),
		).
		file(
			"dir3/cm6.yaml",
			testConfigMap(env.Namespace, "cm6", now),
		)

	deployOptions.FileSystem = inMemFS.toFS()

	// Set up kubeClient
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	require.NoError(t, err, "error creating rest config")

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	require.NoError(t, err, "error creating k8s clientset")

	ctx := context.Background()
	expectedNames := sets.New[string]("cm1", "cm2", "cm3", "cm4", "cm5", "cm6")

	// Test part 1
	testLogger.Section("Deploy(create)", func() {
		err := deployCmd.Execute()
		require.NoError(t, err)

		configMaps, err := kubeClient.CoreV1().ConfigMaps(env.Namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		require.NoError(t, err, "error listing ConfigMaps")

		actualNames := sets.New[string]()
		for _, cm := range configMaps.Items {
			actualNames.Insert(cm.Name)
		}
		require.Empty(t, cmp.Diff(expectedNames, actualNames))
	})

	// Test part 2
	testLogger.Section("Deploy(update)", func() {
		inMemFS.file("dir1/cm7.yaml", testConfigMap(env.Namespace, "cm7", now))
		expectedNames.Insert("cm7")

		err := deployCmd.Execute()
		require.NoError(t, err)

		configMaps, err := kubeClient.CoreV1().ConfigMaps(env.Namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		require.NoError(t, err, "error listing ConfigMaps")

		actualNames := sets.New[string]()
		for _, cm := range configMaps.Items {
			actualNames.Insert(cm.Name)
		}
		require.Empty(t, cmp.Diff(expectedNames, actualNames))
	})

}

type fsBuilder struct {
	mapFS fstest.MapFS
}

func newFSBuilder() *fsBuilder {
	return &fsBuilder{
		mapFS: fstest.MapFS{},
	}
}

func (b *fsBuilder) file(name, contents string) *fsBuilder {
	b.mapFS[name] = &fstest.MapFile{
		Data: []byte(contents),
	}
	return b
}

func (b *fsBuilder) toFS() fstest.MapFS {
	return b.mapFS
}

func testConfigMap(ns, name string, label int64) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: %s
  name: %s
  labels:
    now: "%d"
`, ns, name, label)
}
