// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/k14s/kapp/pkg/kapp/cmd/tools"
)

func AdjustApplyFlags(ssa tools.SSAFlags, af *ApplyFlags) {
	af.ServerSideApply = ssa.SSAEnable
	af.ServerSideForceConflict = ssa.SSAConflict
}
