/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package app

import (
	"fmt"
)

type ExitStatus interface {
	ExitStatus() int
}

type DeployDiffExitStatus struct {
	hasNoChanges bool
}

var _ ExitStatus = DeployDiffExitStatus{}

func (d DeployDiffExitStatus) Error() string {
	numStr := "pending changes"
	if d.hasNoChanges {
		numStr = "no pending changes"
	}
	return fmt.Sprintf("Exiting after diffing with %s (exit status %d)",
		numStr, d.ExitStatus())
}

func (d DeployDiffExitStatus) ExitStatus() int {
	if d.hasNoChanges {
		return 2
	}
	return 3
}
