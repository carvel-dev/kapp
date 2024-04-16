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

func TestKappctrlK14sIoV1alpha1AppFailure(t *testing.T) {
	appTemplate := `
apiVersion: kappctrl.k14s.io/v1alpha1
kind: App
metadata:
  name: test-app
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

	appWithUsefulErrorMessage := fmt.Sprintf(appTemplate, conditionMessage, usefulErrorMessage)
	state := buildKCApp(appWithUsefulErrorMessage, t).IsDoneApplying()
	expectedState := ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    fmt.Sprintf("Reconcile failed:  (message: %s)", usefulErrorMessage),
	}
	require.Equal(t, expectedState, state)

	// Test that kapp falls back to message in condition if usefulErrorMessage is absent
	appWithoutUsefulErrorMessage := fmt.Sprintf(appTemplate, conditionMessage, "")
	state = buildKCApp(appWithoutUsefulErrorMessage, t).IsDoneApplying()
	expectedState = ctlresm.DoneApplyState{
		Done:       true,
		Successful: false,
		Message:    fmt.Sprintf("Reconcile failed:  (message: %s)", conditionMessage),
	}
	require.Equal(t, state, expectedState)

}

func buildKCApp(resourcesBs string, t *testing.T) *ctlresm.KappctrlK14sIoV1alpha1App {
	newResources, err := ctlres.NewFileResource(ctlres.NewBytesSource([]byte(resourcesBs))).Resources()
	require.NoErrorf(t, err, "Expected resources to parse")

	return ctlresm.NewKappctrlK14sIoV1alpha1App(newResources[0])
}
