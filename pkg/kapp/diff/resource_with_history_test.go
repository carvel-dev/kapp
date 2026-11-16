// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff_test

import (
	"testing"

	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestResourceWithoutHistory_DropsEmptyAnnotations(t *testing.T) {
	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  annotations:
    kapp.k14s.io/original-diff-md5: "58e0494c51d30eb3494f7c9198986bb9"
`))

	res, err := ctldiff.NewResourceWithoutHistory(existingRes, nil).Resource()
	require.NoError(t, err)
	require.NotContains(t, res.UnstructuredObject()["metadata"], "annotations")
}

func TestChangeSet_NoUpdateWhenOnlyHistoryAnnotationsWereOnExisting(t *testing.T) {
	newRes := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
`))

	existingRes := ctlres.MustNewResourceFromBytes([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
  annotations:
    kapp.k14s.io/original-diff-md5: "58e0494c51d30eb3494f7c9198986bb9"
`))

	changeFactory := ctldiff.NewChangeFactory(nil, nil, nil, ctldiff.ChangeOpts{})
	changeSet := ctldiff.NewChangeSet([]ctlres.Resource{existingRes}, []ctlres.Resource{newRes},
		ctldiff.ChangeSetOpts{AgainstLastApplied: true}, changeFactory)

	changes, err := changeSet.Calculate()
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.Equal(t, ctldiff.ChangeOpKeep, changes[0].Op())
}
