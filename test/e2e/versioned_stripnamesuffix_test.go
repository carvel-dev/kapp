// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (kapp Kapp, appName string, cleanUp func()) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp = Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	appName = "test-versioned-stripnamesuffix"
	cleanUp = func() {
		kapp.Run([]string{"delete", "-a", appName})
	}

	cleanUp()
	return
}

func testResOverlayPath(name string) string {
	return fmt.Sprintf("res/kustomize/overlays/%s/kapp.yml", name)
}

func kappDeployOverlay(kapp Kapp, name string, app string) (string, error) {
	testResReader, err := TestResReader(testResOverlayPath(name))
	if err != nil {
		return "", fmt.Errorf("Could not load overlay for %s! %w", name, err)
	}
	return kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", app, "--json", "-c"}, RunOpts{IntoNs: true, StdinReader: testResReader})
}

func TestStripNameSuffixBasic(t *testing.T) {
	kapp, appName, cleanup := setup(t)
	defer cleanup()

	if _, err := kappDeployOverlay(kapp, "versioned1", appName); err != nil {
		t.Errorf("Failed to deploy initial overlay!")
	}

	stdout, err := kappDeployOverlay(kapp, "versioned2", appName)

	if err != nil {
		t.Errorf("Failed to deploy next overlay!")
		return
	}

	//fmt.Println(stdout)

	expectedDiff := replaceAnnsLabels(`  ...
  1,  1   data:
  2     -   foo: foo
      2 +   foo: bar
  3,  3   kind: ConfigMap
  4,  4   metadata:
  ...
  7,  7       kapp.k14s.io/app: "1660686583367025336"
  8     -     kapp.k14s.io/association: v1.d0fdf34aa1d77adddf880bb323b33066
      8 +     kapp.k14s.io/association: v1.4fda3fe945a039589026903cb477f5aa
  9,  9     managedFields:
 10, 10     - apiVersion: v1
`)
	expectedNote := "Op:      1 create, 1 delete, 0 update, 0 noop, 0 exists"

	resp := uitest.JSONUIFromBytes(t, []byte(stdout))

	// Ensure the diff is shown
	require.Exactlyf(t, expectedDiff, replaceAnnsLabels(resp.Blocks[0]), "Expected to see correct diff")

	// Ensure old ConfigMap is deleted
	require.Exactlyf(t, expectedNote, replaceAnnsLabels(resp.Tables[0].Notes[0]), "Expected to see correct notes")
}

func TestStripNameSuffixNoop(t *testing.T) {
	kapp, appName, cleanup := setup(t)
	defer cleanup()

	if _, err := kappDeployOverlay(kapp, "versioned1", appName); err != nil {
		t.Errorf("Failed to deploy initial overlay!")
	}

	stdout, err := kappDeployOverlay(kapp, "versioned1", appName)

	if err != nil {
		t.Errorf("Failed to deploy next overlay!")
		return
	}

	expectedNote := "Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists"

	resp := uitest.JSONUIFromBytes(t, []byte(stdout))

	// Ensure current ConfigMap is not deleted
	require.Exactlyf(t, expectedNote, replaceAnnsLabels(resp.Tables[0].Notes[0]), "Expected to see correct notes")
}
