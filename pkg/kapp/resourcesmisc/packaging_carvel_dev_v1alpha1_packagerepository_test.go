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

func TestPackagingCarvelDevV1alpha1PackageRepoFailure(t *testing.T) {
	pkgrTemplate := `
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageRepository
metadata:
  name: test-pkgr
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

	pkgrWithUsefulErrorMessage := fmt.Sprintf(pkgrTemplate, conditionMessage, usefulErrorMessage)
	state := buildKCPkgr(pkgrWithUsefulErrorMessage, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    fmt.Sprintf("Reconcile failed:  (message: %s)", usefulErrorMessage),
	}
	require.Equal(t, expectedState, state)

	// Test that kapp falls back to message in condition if usefulErrorMessage is absent
	pkgrWithoutUsefulErrorMessage := fmt.Sprintf(pkgrTemplate, conditionMessage, "")
	state = buildKCPkgr(pkgrWithoutUsefulErrorMessage, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    fmt.Sprintf("Reconcile failed:  (message: %s)", conditionMessage),
	}
	require.Equal(t, state, expectedState)

}

func buildKCPkgr(resourcesBs string, t *testing.T) *ctlresm.PackagingCarvelDevV1alpha1PackageRepo {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	require.NoErrorf(t, err, "Expected resources to parse")

	return ctlresm.NewPackagingCarvelDevV1alpha1PackageRepo(newResources[0])
}
