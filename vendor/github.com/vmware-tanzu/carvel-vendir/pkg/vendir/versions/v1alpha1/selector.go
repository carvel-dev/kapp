// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fmt"

func HighestConstrainedVersion(versions []string, config VersionSelection) (string, error) {
	switch {
	case config.Semver != nil:
		matchedVers := NewRelaxedSemversNoErr(versions).FilterPrereleases(config.Semver.Prereleases)

		if len(config.Semver.Constraints) > 0 {
			var err error
			matchedVers, err = matchedVers.FilterConstraints(config.Semver.Constraints)
			if err != nil {
				return "", fmt.Errorf("Selecting versions: %s", err)
			}
		}

		highestVersion, found := matchedVers.Highest()
		if !found {
			return "", fmt.Errorf("Expected to find at least one version, but did not")
		}
		return highestVersion, nil

	default:
		return "", fmt.Errorf("Unsupported version selection type (currently supported: semver)")
	}
}
