package diff_test

import (
	"strings"
	"testing"

	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

func TestChangeSetRebaseWithoutNewButWithUnexpectedChanges(t *testing.T) {
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
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	changeFactory := ctldiff.NewChangeFactory(mods)
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{}, changeFactory)

	changes, err := changeSet.Calculate()
	if err != nil {
		t.Fatalf("Expected non-err")
	}

	actualDiff := changes[0].TextDiff().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2       rebased: "1"
  3,  3 -     unexpected: "1"
  4,  3     name: my-res
  5,  4   <---space
`, "<---space", "", -1)

	if actualDiff != expectedDiff {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff, expectedDiff, len(actualDiff), len(expectedDiff))
	}
}

func TestChangeSetRebaseWithNewAndUnexpectedChanges(t *testing.T) {
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
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting},
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew},
		},
		ctlres.FieldCopyMod{
			ResourceMatcher: ctlres.AllResourceMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", "rebased"}),
			Sources:         []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
		},
	}

	changeFactory := ctldiff.NewChangeFactory(mods)
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{}, changeFactory)

	changes, err := changeSet.Calculate()
	if err != nil {
		t.Fatalf("Expected non-err")
	}

	actualDiff := changes[0].TextDiff().FullString()

	expectedDiff := strings.Replace(`  0,  0   metadata:
  1,  1     annotations:
  2,  2 +     new: "1"
  2,  3       rebased: "1"
  3,  4 -     unexpected: "1"
  4,  4     name: my-res
  5,  5   <---space
`, "<---space", "", -1)

	if actualDiff != expectedDiff {
		t.Fatalf("Expected diff to match: actual >>>%s<<< vs expected >>>%s<<< %d %d", actualDiff, expectedDiff, len(actualDiff), len(expectedDiff))
	}
}
