// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"regexp"
	"testing"

	ui "github.com/cppforlife/go-cli-ui/ui"
	uitest "github.com/cppforlife/go-cli-ui/ui/test"
	"github.com/stretchr/testify/require"
)

func run(t *testing.T, overlay1, overlay2 string) (resp ui.JSONUIResp) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	appName := "test-versioned-stripnamesuffix"
	cleanUp := func() {
		kapp.Run([]string{"delete", "-a", appName})
	}

	cleanUp()

	if _, err := kappDeployOverlay(kapp, overlay1, appName); err != nil {
		t.Errorf("Failed to deploy initial overlay!")
	}

	stdout, err := kappDeployOverlay(kapp, overlay2, appName)

	defer cleanUp()

	if err != nil {
		t.Errorf("Failed to deploy next overlay!")
	}

	return uitest.JSONUIFromBytes(t, []byte(stdout))
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

	resp := run(t, "versioned1", "versioned2")

	// comparing the whole diff is unreliable; depending on the k8s version
	// managedFields will (not) be set and as such (not) included in the diff.
	// in that regard it also seems unreliable to assume "data" diff will
	// always be in the first lines.
	addRE := regexp.MustCompile(`(?m)^      \d+ \+   foo: bar$`)
	delRE := regexp.MustCompile(`(?m)^  \d+     \-   foo: foo$`)

	diffBlock := resp.Blocks[0]

	require.Regexpf(t, addRE, diffBlock, "Expected to see new line in diff")
	require.Regexpf(t, delRE, diffBlock, "Expected to see old line in diff")
}

func TestStripNameSuffixDeleteOld(t *testing.T) {

	t.Skip("not yet implemented")

	resp := run(t, "versioned1", "versioned2")

	expectedNote := "Op:      1 create, 1 delete, 0 update, 0 noop, 0 exists"

	actualNote := resp.Tables[0].Notes[0]

	// Ensure old ConfigMap is deleted
	require.Exactlyf(t, expectedNote, actualNote, "Expected to one delete and one create Op")
}

func TestStripNameSuffixNoop(t *testing.T) {

	resp := run(t, "versioned1", "versioned1")

	expectedNote := "Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists"

	actualNote := resp.Tables[0].Notes[0]

	// Ensure current ConfigMap is not deleted
	require.Exactlyf(t, expectedNote, actualNote, "Expected to see no Op's")
}
