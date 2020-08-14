// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
)

type DeployApplyExitStatus struct {
	hasNoChanges bool
}

var _ ExitStatus = DeployApplyExitStatus{}

func (d DeployApplyExitStatus) Error() string {
	numStr := "changes"
	if d.hasNoChanges {
		numStr = "no changes"
	}
	return fmt.Sprintf("Exiting after applying with %s (exit status %d)",
		numStr, d.ExitStatus())
}

func (d DeployApplyExitStatus) ExitStatus() int {
	if d.hasNoChanges {
		return 2
	}
	return 3
}
