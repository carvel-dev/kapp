// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLabeledResourcesCheckDisallowedLabels(t *testing.T) {
	exs := []checkDisallowedLabelsExample{
		{
			Description: "no disallowed labels configured",
			Disallowed:  nil,
		},
		{
			Description: "disallowed label not present on resource",
			Disallowed:  []string{"other"},
		},
		{
			Description: "disallowed label present on resource",
			Disallowed:  []string{"disallowed"},
			ExpectedErr: "Disallowed label errors:\n" +
				"- Resource 'configmap/cm (v1) namespace: ns' has a disallowed label 'disallowed'",
		},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type checkDisallowedLabelsExample struct {
	Description string
	Disallowed  []string
	ExpectedErr string
}

func (e checkDisallowedLabelsExample) Check(t *testing.T) {
	res := MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  namespace: ns
  labels:
    allowed: "1"
    disallowed: "1"
`))

	err := (&LabeledResources{}).checkDisallowedLabels([]Resource{res}, e.Disallowed)
	if len(e.ExpectedErr) > 0 {
		require.EqualError(t, err, e.ExpectedErr, e.Description)
		return
	}

	require.NoError(t, err, e.Description)
}
