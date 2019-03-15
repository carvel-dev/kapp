package resources_test

import (
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

func TestModFieldCopy(t *testing.T) {
	exs := []modFieldCopyExample{
		{
			Description: "copy from new, when existing has it",
			Res: `
metadata:
  labels: {}`,
			Expected: `
metadata:
  labels:
    label-key: new-label-val`,
			Sources: []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
			Path:    ctlres.NewPathFromStrings([]string{"metadata", "labels", "label-key"}),
			NewRes: `
metadata:
  labels:
    label-key: new-label-val`,
			ExistingRes: `
metadata:
  labels:
    label-key: existing-label-val`,
		},
		{
			Description: "copy from existing, when new has it",
			Res: `
metadata:
  labels: {}`,
			Expected: `
metadata:
  labels:
    label-key: existing-label-val`,
			Sources: []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceExisting, ctlres.FieldCopyModSourceNew},
			Path:    ctlres.NewPathFromStrings([]string{"metadata", "labels", "label-key"}),
			NewRes: `
metadata:
  labels:
    label-key: new-label-val`,
			ExistingRes: `
metadata:
  labels:
    label-key: existing-label-val`,
		},
		{
			Description: "fall back to existing when new doesnt have it",
			Res: `
metadata:
  labels: {}`,
			Expected: `
metadata:
  labels:
    label-key: existing-label-val`,
			Sources: []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
			Path:    ctlres.NewPathFromStrings([]string{"metadata", "labels", "label-key"}),
			NewRes: `
metadata:
  labels: {}`,
			ExistingRes: `
metadata:
  labels:
    label-key: existing-label-val`,
		},
		{
			Description: "does not set if does not exist in new or existing",
			Res: `
metadata:
  labels: {}`,
			Expected: `
metadata:
  labels: {}`,
			Sources: []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
			Path:    ctlres.NewPathFromStrings([]string{"metadata", "labels", "label-key"}),
			NewRes: `
metadata:
  labels: {}`,
			ExistingRes: `
metadata:
  labels:
    another-label-key: existing-label-val`,
		},
		{
			Description: "copies value nested in an array",
			Res: `
metadata:
  labels:
  - {}`,
			Expected: `
metadata:
  labels:
  - label-key: existing-label-val`,
			Sources: []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
			Path: ctlres.Path{
				ctlres.NewPathPartFromString("metadata"),
				ctlres.NewPathPartFromString("labels"),
				ctlres.NewPathPartFromIndexAll(),
				ctlres.NewPathPartFromString("label-key"),
			},
			NewRes: `
metadata:
  labels:
  - {}`,
			ExistingRes: `
metadata:
  labels:
  - label-key: existing-label-val`,
		},
		{
			Description: "copies value nested in an array with multiple items",
			Res: `
metadata:
  labels:
  - {}
  - {}
  - label-key: already-preset`,
			Expected: `
metadata:
  labels:
  - label-key: existing-label-val
  - label-key: new-label-val
  - label-key: already-preset`,
			Sources: []ctlres.FieldCopyModSource{ctlres.FieldCopyModSourceNew, ctlres.FieldCopyModSourceExisting},
			Path: ctlres.Path{
				ctlres.NewPathPartFromString("metadata"),
				ctlres.NewPathPartFromString("labels"),
				ctlres.NewPathPartFromIndexAll(),
				ctlres.NewPathPartFromString("label-key"),
			},
			NewRes: `
metadata:
  labels:
  - {}
  - label-key: new-label-val`,
			ExistingRes: `
metadata:
  labels:
  - label-key: existing-label-val
  - label-key: another-existing-label-val`,
		},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type modFieldCopyExample struct {
	Description string
	Res         string
	Path        ctlres.Path
	Sources     []ctlres.FieldCopyModSource
	NewRes      string
	ExistingRes string
	Expected    string
}

func (e modFieldCopyExample) Check(t *testing.T) {
	res := ctlres.MustNewResourceFromBytes([]byte(e.Res))

	ress := map[ctlres.FieldCopyModSource]ctlres.Resource{}
	if len(e.NewRes) > 0 {
		ress[ctlres.FieldCopyModSourceNew] = ctlres.MustNewResourceFromBytes([]byte(e.NewRes))
	}
	if len(e.ExistingRes) > 0 {
		ress[ctlres.FieldCopyModSourceExisting] = ctlres.MustNewResourceFromBytes([]byte(e.ExistingRes))
	}

	err := ctlres.FieldCopyMod{
		ResourceMatcher: ctlres.AllResourceMatcher{},
		Path:            e.Path,
		Sources:         e.Sources,
	}.ApplyFromMultiple(res, ress)
	if err != nil {
		t.Fatalf("Expected no err, but was %s", err)
	}

	resultBs, err := res.AsYAMLBytes()
	if err != nil {
		t.Fatalf("Expected no err, but was %s", err)
	}

	expectEqualsStripped(t, e.Description, string(resultBs), e.Expected)
}
