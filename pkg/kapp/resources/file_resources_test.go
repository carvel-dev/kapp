// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"testing"
	"testing/fstest"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/stretchr/testify/require"
)

func TestFileResourceResourcesSetsOrigin(t *testing.T) {
	fsys := fstest.MapFS{
		"config/app.yml": &fstest.MapFile{Data: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-2
`)},
	}

	fileRs, err := ctlres.NewFileResources(fsys, "config")
	require.NoError(t, err)
	require.Len(t, fileRs, 1)

	rs, err := fileRs[0].Resources()
	require.NoError(t, err)
	require.Len(t, rs, 2)
	require.Equal(t, "file 'config/app.yml' doc 1", rs[0].Origin())
	require.Equal(t, "file 'config/app.yml' doc 2", rs[1].Origin())
}

func TestFileResourceResourcesReportsOriginOfInvalidDoc(t *testing.T) {
	fsys := fstest.MapFS{
		"config/a.yml": &fstest.MapFile{Data: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
`)},
		"config/b.yml": &fstest.MapFile{Data: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-2
---
not a resource
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-3
`)},
	}

	fileRs, err := ctlres.NewFileResources(fsys, "config")
	require.NoError(t, err)
	require.Len(t, fileRs, 2)

	_, err = fileRs[0].Resources()
	require.NoError(t, err)

	_, err = fileRs[1].Resources()
	require.ErrorContains(t, err, "Parsing file 'config/b.yml' doc 2: error unmarshaling JSON: ")

	_, err = ctlres.NewFileResource(ctlres.NewBytesSource([]byte("key: value\n---\nkey: [\n"))).Resources()
	require.ErrorContains(t, err, "Parsing bytes doc 2: error converting YAML to JSON: ")
}
