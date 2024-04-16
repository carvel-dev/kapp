// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"testing"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestChangeSet_ExistingVersioned_NewNonVersioned_Resource(t *testing.T) {
	newRs := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret-ver-1
  annotations:
    kapp.k14s.io/versioned: ""
`))

	changeSetWithVerRes := NewChangeSetWithVersionedRs([]ctlres.Resource{existingRes}, []ctlres.Resource{newRs}, nil,
		ChangeSetOpts{}, ChangeFactory{})

	changes, err := changeSetWithVerRes.Calculate()
	require.NoError(t, err)

	require.Len(t, changes, 2)

	require.Equal(t, ChangeOpDelete, changes[0].Op(), "Expected to get deleted")

	require.Equal(t, ChangeOpAdd, changes[1].Op(), "Expected to get added")

	expectedDiff1 := `  0,  0 - apiVersion: v1
  1,  0 - kind: Secret
  2,  0 - metadata:
  3,  0 -   annotations:
  4,  0 -     kapp.k14s.io/versioned: ""
  5,  0 -   name: secret-ver-1
  6,  0 - 
`
	checkChangeDiff(t, changes[0], expectedDiff1)

	expectedDiff2 := `  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   name: secret
  0,  4 + 
`
	checkChangeDiff(t, changes[1], expectedDiff2)
}

func TestChangeSet_ExistingNonVersioned_NewVersioned_Resource(t *testing.T) {
	newRs := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
`))

	changeSetWithVerRes := NewChangeSetWithVersionedRs([]ctlres.Resource{existingRes}, []ctlres.Resource{newRs}, nil,
		ChangeSetOpts{}, ChangeFactory{})

	changes, err := changeSetWithVerRes.Calculate()
	require.NoError(t, err)

	require.Len(t, changes, 2)

	require.Equal(t, ChangeOpAdd, changes[0].Op(), "Expected to get added")

	require.Equal(t, ChangeOpDelete, changes[1].Op(), "Expected to get deleted")

	expectedDiff1 := `  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +   name: secret-ver-1
  0,  6 + 
`
	checkChangeDiff(t, changes[0], expectedDiff1)

	expectedDiff2 := `  0,  0 - apiVersion: v1
  1,  0 - kind: Secret
  2,  0 - metadata:
  3,  0 -   name: secret
  4,  0 - 
`
	checkChangeDiff(t, changes[1], expectedDiff2)
}

func TestChangeSet_ExistingNonVersioned_NewVersioneKeepOrg_Resource(t *testing.T) {
	newRs := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
  annotations:
    kapp.k14s.io/versioned: ""
    kapp.k14s.io/versioned-keep-original: ""
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: Secret
metadata:
  name: secret
`))

	changeSetWithVerRes := NewChangeSetWithVersionedRs([]ctlres.Resource{existingRes}, []ctlres.Resource{newRs}, nil,
		ChangeSetOpts{}, ChangeFactory{})

	changes, err := changeSetWithVerRes.Calculate()
	require.NoError(t, err)

	require.Len(t, changes, 2)

	require.Equal(t, ChangeOpAdd, changes[0].Op(), "Expected to get added")

	require.Equal(t, ChangeOpUpdate, changes[1].Op(), "Expected to get updated")

	expectedDiff1 := `  0,  0 + apiVersion: v1
  0,  1 + kind: Secret
  0,  2 + metadata:
  0,  3 +   annotations:
  0,  4 +     kapp.k14s.io/versioned: ""
  0,  5 +     kapp.k14s.io/versioned-keep-original: ""
  0,  6 +   name: secret-ver-1
  0,  7 + 
`
	checkChangeDiff(t, changes[0], expectedDiff1)

	expectedDiff2 := `  0,  0   apiVersion: v1
  1,  1   kind: Secret
  2,  2   metadata:
  3,  3 +   annotations:
  3,  4 +     kapp.k14s.io/versioned: ""
  3,  5 +     kapp.k14s.io/versioned-keep-original: ""
  3,  6     name: secret
  4,  7   
`
	checkChangeDiff(t, changes[1], expectedDiff2)
}

func checkChangeDiff(t *testing.T, change Change, expectedDiff string) {
	actualDiffString := change.ConfigurableTextDiff().Full().FullString()

	require.Equal(t, expectedDiff, actualDiffString, "Expected diff to match")
}
