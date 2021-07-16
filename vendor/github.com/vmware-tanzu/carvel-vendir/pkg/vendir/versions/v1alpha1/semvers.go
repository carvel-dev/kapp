// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"sort"
	"strings"

	semver "github.com/k14s/semver/v4"
)

type Semvers struct {
	versions []SemverWrap
}

type SemverWrap struct {
	semver.Version
	Original string
}

func NewSemver(version string) (SemverWrap, error) {
	parsedVersion, err := semver.Parse(version)
	if err != nil {
		return SemverWrap{}, err
	}

	return SemverWrap{parsedVersion, version}, nil
}

func NewRelaxedSemver(version string) (SemverWrap, error) {
	parsableVersion := version
	if strings.HasPrefix(version, "v") {
		parsableVersion = strings.TrimPrefix(version, "v")
	}

	parsedVersion, err := semver.Parse(parsableVersion)
	if err != nil {
		return SemverWrap{}, err
	}

	return SemverWrap{parsedVersion, version}, nil
}

func NewRelaxedSemversNoErr(versions []string) Semvers {
	var parsedVersions []SemverWrap

	for _, vStr := range versions {
		ver, err := NewRelaxedSemver(vStr)
		if err != nil {
			continue
		}
		parsedVersions = append(parsedVersions, ver)
	}

	return Semvers{parsedVersions}
}

func (v Semvers) Sorted() Semvers {
	var versions []SemverWrap

	for _, ver := range v.versions {
		versions = append(versions, ver)
	}

	sort.SliceStable(versions, func(i, j int) bool {
		return versions[i].Version.LT(versions[j].Version)
	})

	return Semvers{versions}
}

func (v Semvers) FilterConstraints(constraintList string) (Semvers, error) {
	constraints, err := semver.ParseRange(constraintList)
	if err != nil {
		return Semvers{}, fmt.Errorf("Parsing version constraint '%s': %s", constraintList, err)
	}

	var matchingVersions []SemverWrap

	for _, ver := range v.versions {
		if constraints(ver.Version) {
			matchingVersions = append(matchingVersions, ver)
		}
	}

	return Semvers{matchingVersions}, nil
}

func (v Semvers) FilterPrereleases(prereleases *VersionSelectionSemverPrereleases) Semvers {
	if prereleases == nil {
		// Exclude all prereleases
		var result []SemverWrap
		for _, ver := range v.versions {
			if len(ver.Version.Pre) == 0 {
				result = append(result, ver)
			}
		}
		return Semvers{result}
	}

	preIdentifiersAsMap := prereleases.IdentifiersAsMap()

	var result []SemverWrap
	for _, ver := range v.versions {
		if len(ver.Version.Pre) == 0 || v.shouldKeepPrerelease(ver.Version, preIdentifiersAsMap) {
			result = append(result, ver)
		}
	}
	return Semvers{result}
}

func (Semvers) shouldKeepPrerelease(ver semver.Version, preIdentifiersAsMap map[string]struct{}) bool {
	if len(preIdentifiersAsMap) == 0 {
		return true
	}
	for _, prePart := range ver.Pre {
		if len(prePart.VersionStr) > 0 {
			if _, found := preIdentifiersAsMap[prePart.VersionStr]; found {
				return true
			}
		}
	}
	return false
}

func (v Semvers) Highest() (string, bool) {
	v = v.Sorted()

	if len(v.versions) == 0 {
		return "", false
	}

	return v.versions[len(v.versions)-1].Original, true
}

func (v Semvers) All() []string {
	var verStrs []string
	for _, ver := range v.versions {
		verStrs = append(verStrs, ver.Original)
	}
	return verStrs
}
