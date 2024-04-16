// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"os"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

const (
	appliedResAnnKey        = "kapp.k14s.io/original"
	appliedResDiffMD5AnnKey = "kapp.k14s.io/original-diff-md5"

	// Following fields useful for debugging:
	debugAppliedResDiffAnnKey     = "kapp.k14s.io/original-diff"
	debugAppliedResDiffFullAnnKey = "kapp.k14s.io/original-diff-full"

	disableOriginalAnnKey = "kapp.k14s.io/disable-original"
)

var (
	resourceWithHistoryDebug = os.Getenv("KAPP_DEBUG_RESOURCE_WITH_HISTORY") == "true"
)

type ResourceWithHistory struct {
	resource                                 ctlres.Resource
	changeFactory                            *ChangeFactory
	diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod
}

func NewResourceWithHistory(resource ctlres.Resource,
	changeFactory *ChangeFactory, diffAgainstLastAppliedFieldExclusionMods []ctlres.FieldRemoveMod) ResourceWithHistory {

	return ResourceWithHistory{resource.DeepCopy(), changeFactory, diffAgainstLastAppliedFieldExclusionMods}
}
func NewResourceWithoutHistory(resource ctlres.Resource, fieldExclusionMods []ctlres.FieldRemoveMod) ResourceWithoutHistory {
	return ResourceWithoutHistory{res: resource, fieldExclusionMods: fieldExclusionMods}
}

// LastAppliedResource will return "last applied" resource that was saved
// iff it still matches actually saved resource on the cluster (noted at the time of saving).
func (r ResourceWithHistory) LastAppliedResource() ctlres.Resource {
	recalculatedLastAppliedChanges, expectedDiffMD5, expectedDiff := r.recalculateLastAppliedChange()

	for _, recalculatedLastAppliedChange := range recalculatedLastAppliedChanges {
		md5Matches := recalculatedLastAppliedChange.OpsDiff().MinimalMD5() == expectedDiffMD5

		if resourceWithHistoryDebug {
			fmt.Printf("%s: md5 matches (%t) prev %s recalc %s\n----> pref diff\n%s\n----> recalc diff\n%s\n",
				r.resource.Description(), md5Matches,
				expectedDiffMD5, recalculatedLastAppliedChange.OpsDiff().MinimalMD5(),
				expectedDiff, recalculatedLastAppliedChange.OpsDiff().MinimalString())
		}

		if md5Matches {
			return recalculatedLastAppliedChange.AppliedResource()
		}
	}

	return nil
}

func (r ResourceWithHistory) AllowsRecordingLastApplied() bool {
	_, found := r.resource.Annotations()[disableOriginalAnnKey]
	return !found
}

func (r ResourceWithHistory) RecordLastAppliedResource(appliedChange Change) (ctlres.Resource, bool, error) {
	// Use compact representation to take as little space as possible
	// because annotation value max length is 262144 characters
	// (https://github.com/carvel-dev/kapp/issues/48).
	appliedResBytes, err := appliedChange.AppliedResource().AsCompactBytes()
	if err != nil {
		return nil, true, err
	}

	diff := appliedChange.OpsDiff()

	if resourceWithHistoryDebug {
		fmt.Printf("%s: recording md5 %s\n---> \n%s\n",
			r.resource.Description(), diff.MinimalMD5(), diff.MinimalString())
	}

	annsMod := ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs: map[string]string{
			appliedResAnnKey:        string(appliedResBytes),
			appliedResDiffMD5AnnKey: diff.MinimalMD5(),

			// Following fields useful for debugging:
			//   debugAppliedResDiffAnnKey:     diff.MinimalString(),
			//   debugAppliedResDiffFullAnnKey: diff.FullString(),
		},
	}

	const annValMaxLen = 262144

	// kapp deploy should work without adding disable annotation when annotation value max length exceed
	// (https://github.com/carvel-dev/kapp/issues/410)
	for _, annVal := range annsMod.KVs {
		if len(annVal) > annValMaxLen {
			return nil, false, nil
		}
	}

	resultRes := r.resource.DeepCopy()

	err = annsMod.Apply(resultRes)
	if err != nil {
		return nil, true, err
	}

	return resultRes, true, nil
}

func (r ResourceWithHistory) CalculateChange(appliedRes ctlres.Resource) (Change, error) {
	// Remove fields specified to be excluded (as they may be generated
	// by the server, hence would be racy to be rebased)
	removeMods := r.diffAgainstLastAppliedFieldExclusionMods

	existingRes, err := NewResourceWithRemovedFields(r.resource, removeMods).Resource()
	if err != nil {
		return nil, err
	}

	return r.newExactHistorylessChange(existingRes, appliedRes)
}

func (r ResourceWithHistory) recalculateLastAppliedChange() ([]Change, string, string) {
	lastAppliedResBytes := r.resource.Annotations()[appliedResAnnKey]
	lastAppliedDiffMD5 := r.resource.Annotations()[appliedResDiffMD5AnnKey]

	if len(lastAppliedResBytes) == 0 || len(lastAppliedDiffMD5) == 0 {
		return nil, "", ""
	}

	lastAppliedRes, err := ctlres.NewResourceFromBytes([]byte(lastAppliedResBytes))
	if err != nil {
		return nil, "", ""
	}

	recalculatedChange, err := r.CalculateChange(lastAppliedRes)
	if err != nil {
		return nil, "", "" // TODO deal with error?
	}

	lastAppliedDiff := r.resource.Annotations()[debugAppliedResDiffAnnKey]

	return []Change{recalculatedChange}, lastAppliedDiffMD5, lastAppliedDiff
}

func (r ResourceWithHistory) newExactHistorylessChange(existingRes, newRes ctlres.Resource) (Change, error) {
	// If annotations are not removed line numbers will be mismatched
	existingRes, err := ResourceWithoutHistory{existingRes, nil}.Resource()
	if err != nil {
		return nil, err
	}

	newRes, err = ResourceWithoutHistory{newRes, nil}.Resource()
	if err != nil {
		return nil, err
	}

	return r.changeFactory.NewExactChange(existingRes, newRes)
}

type ResourceWithoutHistory struct {
	res                ctlres.Resource
	fieldExclusionMods []ctlres.FieldRemoveMod
}

func (r ResourceWithoutHistory) Resource() (ctlres.Resource, error) {
	res := r.res.DeepCopy()

	for _, t := range r.removeAppliedResAnnKeysMods() {
		err := t.Apply(res)
		if err != nil {
			return nil, err
		}
	}

	for _, t := range r.fieldExclusionMods {
		err := t.Apply(res)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (ResourceWithoutHistory) removeAppliedResAnnKeysMods() []ctlres.ResourceMod {
	return []ctlres.ResourceMod{
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", appliedResAnnKey}),
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", debugAppliedResDiffAnnKey}),
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", appliedResDiffMD5AnnKey}),
		},
		ctlres.FieldRemoveMod{
			ResourceMatcher: ctlres.AllMatcher{},
			Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations", debugAppliedResDiffFullAnnKey}),
		},
	}
}
