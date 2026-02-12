// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diffgraph_test

import (
	"testing"

	ctldgraph "carvel.dev/kapp/pkg/kapp/diffgraph"
	"github.com/stretchr/testify/require"
)

func TestNewChangeGroupFromAnnString(t *testing.T) {
	names := []string{
		"valid",
		"valid-name",
		"valid/valid",
		"valid-name/valid",
		"valid-name.com",
		"valid-name.com/valid",
		"valid-name.com/valid-name_Another_Name--valid",
		"valid-name.com/valid-name_CustomResourceDefinition--valid",
		// Example from pinniped of a long name
		"change-groups.kapp.k14s.io/crds-authentication.concierge.pinniped.dev-WebhookAuthenticator",
	}
	for _, name := range names {
		cg, err := ctldgraph.NewChangeGroupFromAnnString(name)
		require.NoError(t, err)
		require.Equal(t, name, cg.Name)
	}

	names = []string{
		"_",
		"invalid/",
		"invalid/_",
		"/_",
	}
	for _, name := range names {
		_, err := ctldgraph.NewChangeGroupFromAnnString(name)
		require.Error(t, err)
	}
}
