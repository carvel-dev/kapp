// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc_test

import (
	"fmt"
	"testing"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
	"github.com/stretchr/testify/require"
)

func TestPackagingCarvelDevV1alpha1PackageInstallFailure(t *testing.T) {
	pkgiTemplate := `
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
  name: test-pkgi
  generation: 1
status:
  observedGeneration: 1
  conditions:
  - message: %s
    status: "True"
    type: ReconcileFailed
  usefulErrorMessage: %s
`

	conditionMessage := "Truncated error message"
	usefulErrorMessage := "Detailed error message"

	pkgiWithUsefulErrorMessage := fmt.Sprintf(pkgiTemplate, conditionMessage, usefulErrorMessage)
	state := buildKCPkgi(pkgiWithUsefulErrorMessage, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    fmt.Sprintf("Reconcile failed: message: %s", usefulErrorMessage),
	}
	require.Equal(t, expectedState, state)

	// Test that kapp falls back to message in condition if usefulErrorMessage is absent
	pkgiWithoutUsefulErrorMessage := fmt.Sprintf(pkgiTemplate, conditionMessage, "")
	state = buildKCPkgi(pkgiWithoutUsefulErrorMessage, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    fmt.Sprintf("Reconcile failed: message: %s", conditionMessage),
	}
	require.Equal(t, state, expectedState)

}

func buildKCPkgi(resourcesBs string, t *testing.T) *ctlresm.PackagingCarvelDevV1alpha1PackageInstall {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	require.NoErrorf(t, err, "Expected resources to parse")

	return ctlresm.NewPackagingCarvelDevV1alpha1PackageInstall(newResources[0])
}
