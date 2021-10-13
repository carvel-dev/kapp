// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"testing"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestModStringMapAppend(t *testing.T) {
	exs := []modStringMapAppendExample{
		{
			Description: "append leaf key that exists",
			Res: `
metadata:
  labels:
    label-key: label-val`,
			Expected: `
metadata:
  labels:
    label-key: new-label-val`,
			KVs:  map[string]string{"label-key": "new-label-val"},
			Path: ctlres.NewPathFromStrings([]string{"metadata", "labels"}),
		},
		{
			Description: "append leaf key that does not exist",
			Res: `
metadata:
  labels: {}`,
			Expected: `
metadata:
  labels:
    label-key: new-label-val`,
			KVs:  map[string]string{"label-key": "new-label-val"},
			Path: ctlres.NewPathFromStrings([]string{"metadata", "labels"}),
		},
		{
			Description: "append parent key that exists",
			Res: `
metadata:
  labels:
    label-key: label-val`,
			Expected: `
metadata:
  labels:
    label-key: label-val
  new-labels: new-labels`,
			KVs:  map[string]string{"new-labels": "new-labels"},
			Path: ctlres.NewPathFromStrings([]string{"metadata"}),
		},
		{
			Description: "append parent key that does not exist",
			Res: `
metadata:
  labels:
    label-key: label-val`,
			Expected: `
metadata:
  labels:
    label-key: label-val
  not-labels:
    new-labels: new-labels`,
			KVs:  map[string]string{"new-labels": "new-labels"},
			Path: ctlres.NewPathFromStrings([]string{"metadata", "not-labels"}),
		},
		{
			Description: "append leaf key that exists under array",
			Res: `
metadata:
  labels:
  - label-key: label-val`,
			Expected: `
metadata:
  labels:
  - label-key: label-val
    new-label-key: new-label-val`,
			KVs: map[string]string{"new-label-key": "new-label-val"},
			Path: ctlres.Path{
				ctlres.NewPathPartFromString("metadata"),
				ctlres.NewPathPartFromString("labels"),
				ctlres.NewPathPartFromIndex(0),
			},
		},
		{
			Description: "append leaf key that does not exist under array",
			Res: `
metadata:
  labels:
  - label-key: label-val`,
			Expected: `
metadata:
  labels:
  - label-key: label-val
    not-labels:
      new-label-key: new-label-val`,
			KVs: map[string]string{"new-label-key": "new-label-val"},
			Path: ctlres.Path{
				ctlres.NewPathPartFromString("metadata"),
				ctlres.NewPathPartFromString("labels"),
				ctlres.NewPathPartFromIndex(0),
				ctlres.NewPathPartFromString("not-labels"),
			},
		},
		{
			Description: "append multiple keys that do or do not exist under array",
			Res: `
metadata:
  labels:
  - label-key: label-val
  - not-label-key: label-val`,
			Expected: `
metadata:
  labels:
  - label-key: label-val
    not-labels:
      new-label-key: new-label-val
  - not-label-key: label-val
    not-labels:
      new-label-key: new-label-val`,
			KVs: map[string]string{"new-label-key": "new-label-val"},
			Path: ctlres.Path{
				ctlres.NewPathPartFromString("metadata"),
				ctlres.NewPathPartFromString("labels"),
				ctlres.NewPathPartFromIndexAll(),
				ctlres.NewPathPartFromString("not-labels"),
			},
		},
		{
			Description: "append multiple keys that do or do not exist under array with array following on",
			Res: `
metadata:
  labels:
  - sub-labels:
    - label-key: label-val
  - {}`,
			Expected: `
metadata:
  labels:
  - sub-labels:
    - label-key: label-val
      new-label-key: new-label-val
  - {}`,
			KVs: map[string]string{"new-label-key": "new-label-val"},
			Path: ctlres.Path{
				ctlres.NewPathPartFromString("metadata"),
				ctlres.NewPathPartFromString("labels"),
				ctlres.NewPathPartFromIndexAll(),
				ctlres.NewPathPartFromString("sub-labels"),
				ctlres.NewPathPartFromIndexAll(),
			},
		},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type modStringMapAppendExample struct {
	Description string
	Res         string
	Path        ctlres.Path
	KVs         map[string]string
	Expected    string
}

func (e modStringMapAppendExample) Check(t *testing.T) {
	res, err := ctlres.NewResourceFromBytes([]byte(e.Res))
	require.NoError(t, err)

	err = ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            e.Path,
		KVs:             e.KVs,
	}.Apply(res)
	require.NoError(t, err)

	resultBs, err := res.AsYAMLBytes()
	require.NoError(t, err)

	expectEqualsStripped(t, e.Description, string(resultBs), e.Expected)
}
