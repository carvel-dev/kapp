// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff_test

import (
	"strings"
	"testing"

	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestChangeSet_RebaseWithoutNew_And_WithUnexpectedChanges(t *testing.T) {
	newRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
    unexpected: "1"
    rebased: "1"
`))

	mods := []ctlres.ResourceModWithMultiple{
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	changeFactory := ctldiff.NewChangeFactory(mods, nil, nil, ctldiff.ChangeOpts{false})
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{}, changeFactory)

	changes, err := changeSet.Calculate()
	require.NoError(t, err)

	actualDiff := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2       rebased: "1"
  3,  3 -     unexpected: "1"
  4,  3     name: my-res
  5,  4   <---space
`, "<---space", "", -1)

	require.Equal(t, expectedDiff, actualDiff, "Expected diff to match")
}

func TestChangeSet_RebaseWithNew_And_UnexpectedChanges(t *testing.T) {
	newRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
    new: "1"
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
    unexpected: "1"
    rebased: "1"
`))

	mods := []ctlres.ResourceModWithMultiple{
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	changeFactory := ctldiff.NewChangeFactory(mods, nil, nil, ctldiff.ChangeOpts{false})
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{}, changeFactory)

	changes, err := changeSet.Calculate()
	require.NoError(t, err)

	actualDiff := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2 +     new: "1"
  2,  3       rebased: "1"
  3,  4 -     unexpected: "1"
  4,  4     name: my-res
  5,  5   <---space
`, "<---space", "", -1)

	require.Equal(t, expectedDiff, actualDiff, "Expected diff to match")
}

func TestChangeSet_WithoutNew_And_WithoutUnexpectedChanges_And_IgnoredFields(t *testing.T) {
	newRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
    unexpected-ignored: "1"
    rebased: "1"
    kapp.k14s.io/original: |
      metadata:
        name: my-res
    kapp.k14s.io/original-diff-md5: "58e0494c51d30eb3494f7c9198986bb9"
`))

	rebaseMods := []ctlres.ResourceModWithMultiple{
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	ignoreFieldsMods := []ctlres.FieldRemoveMod{
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "unexpected-ignored"}),
		},
	}

	changeFactory := ctldiff.NewChangeFactory(rebaseMods, ignoreFieldsMods, nil, ctldiff.ChangeOpts{false})
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{AgainstLastApplied: true}, changeFactory)

	changes, err := changeSet.Calculate()
	require.NoError(t, err)

	actualDiff := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2       rebased: "1"
  3,  3     name: my-res
  4,  4   <---space
`, "<---space", "", -1)

	require.Equal(t, expectedDiff, actualDiff, "Expected diff to match")
}

func TestChangeSet_WithoutNew_And_WithUnexpectedChanges_And_IgnoredFields(t *testing.T) {
	newRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
    unexpected: "1"
    rebased-ignored: "1"
    rebased: "1"
    kapp.k14s.io/original: |
      metadata:
        name: my-res
    kapp.k14s.io/original-diff-md5: "58e0494c51d30eb3494f7c9198986bb9"
`))

	rebaseMods := []ctlres.ResourceModWithMultiple{
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased-ignored"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	ignoreFieldsMods := []ctlres.FieldRemoveMod{
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased-ignored"}),
		},
	}

	changeFactory := ctldiff.NewChangeFactory(rebaseMods, ignoreFieldsMods, nil, ctldiff.ChangeOpts{false})
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{AgainstLastApplied: true}, changeFactory)

	changes, err := changeSet.Calculate()
	require.NoError(t, err)

	actualDiff := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2       rebased: "1"
  3,  3       rebased-ignored: "1"
  4,  4 -     unexpected: "1"
  5,  4     name: my-res
  6,  5   <---space
`, "<---space", "", -1)

	require.Equal(t, expectedDiff, actualDiff, "Expected diff to match")
}

func TestRebaseWithNilNewAndUnexpectedChanges(t *testing.T) {
	newRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
metadata:
  name: my-res
  annotations:
    unexpected: "1"
    rebased: "1"
`))

	mods := []ctlres.ResourceModWithMultiple{
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	changeFactory := ctldiff.NewChangeFactory(mods, nil, nil, ctldiff.ChangeOpts{false})
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{}, changeFactory)

	changes, err := changeSet.Calculate()
	require.NoError(t, err)

	actualDiff := changes[0].ConfigurableTextDiff().Full().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2       rebased: "1"
  3,  3 -     unexpected: "1"
  4,  3     name: my-res
  5,  4   <---space
`, "<---space", "", -1)

	require.Equal(t, expectedDiff, actualDiff, "Expected diff to match")
}
