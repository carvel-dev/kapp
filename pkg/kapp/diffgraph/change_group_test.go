// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph_test

import (
	"strings"
	"testing"

	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
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
		// Allow arbitrary long names since it might be populated with data via placeholders
		"valid-name.com/valid-name_CustomResourceDefinition--valid" + strings.Repeat("a", 1000),
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
