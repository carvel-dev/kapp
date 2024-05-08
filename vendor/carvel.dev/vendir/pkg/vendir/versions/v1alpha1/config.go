// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"bytes"
	"encoding/json"
	"strings"
)

// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type VersionSelection struct {
	Semver *VersionSelectionSemver `json:"semver,omitempty" protobuf:"bytes,1,opt,name=semver"`
}

// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type VersionSelectionSemver struct {
	Constraints string                             `json:"constraints,omitempty" protobuf:"bytes,1,opt,name=constraints"`
	Prereleases *VersionSelectionSemverPrereleases `json:"prereleases,omitempty" protobuf:"bytes,2,opt,name=prereleases"`
}

// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type VersionSelectionSemverPrereleases struct {
	Identifiers []string `json:"identifiers,omitempty" protobuf:"bytes,1,opt,name=identifiers"`
}

func (p VersionSelectionSemverPrereleases) IdentifiersAsMap() map[string]struct{} {
	result := map[string]struct{}{}
	for _, name := range p.Identifiers {
		result[name] = struct{}{}
	}
	return result
}

func (vs VersionSelection) Description() string {
	// json.Marshal encodes <,>,& as unicode replacement runes
	// (https://pkg.go.dev/encoding/json#Marshal)
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(vs); err != nil {
		return "?"
	}
	return strings.TrimSpace(string(buffer.Bytes()))
}
