// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package website

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Files map is modified by ./generated.go created during ./hack/build.sh
var Files = map[string]File{}
