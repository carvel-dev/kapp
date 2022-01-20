// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/k14s/kapp/pkg/kapp/cmd/tools/ssa"
)

func AdjustApplyFlags(ssa ssa.SSAFlags, af *ApplyFlags) {
	af.ServerSideApply = ssa.Enabled
	af.ServerSideForceConflict = ssa.ForceConflict
	af.FieldManagerName = ssa.FieldManagerName
}
