// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// +k8s:deepcopy-gen=true
type VersionSelection struct {
	Semver *VersionSelectionSemver `json:"semver,omitempty"`
}

// +k8s:deepcopy-gen=true
type VersionSelectionSemver struct {
	Constraints string                             `json:"constraints,omitempty"`
	Prereleases *VersionSelectionSemverPrereleases `json:"prereleases,omitempty"`
}

// +k8s:deepcopy-gen=true
type VersionSelectionSemverPrereleases struct {
	Identifiers []string `json:"identifiers,omitempty"`
}

func (p VersionSelectionSemverPrereleases) IdentifiersAsMap() map[string]struct{} {
	result := map[string]struct{}{}
	for _, name := range p.Identifiers {
		result[name] = struct{}{}
	}
	return result
}
