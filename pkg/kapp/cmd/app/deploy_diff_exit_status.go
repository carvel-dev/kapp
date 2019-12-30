package app

import (
	"fmt"
)

type DeployDiffExitStatus struct {
	hasNoChanges bool
}

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
