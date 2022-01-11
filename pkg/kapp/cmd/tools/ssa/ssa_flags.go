// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package ssa

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

type SSAFlags struct {
	SSAEnable        bool
	SSAForceConflict bool
	FieldManagerName string
}

func (s *SSAFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.SSAEnable, "ssa-enable", false, "Use server side apply")
	cmd.Flags().StringVar(&s.FieldManagerName, "ssa-field-manager", "kapp-server-side-apply", "Name of the manager used to track field ownership")
	cmd.Flags().BoolVar(&s.SSAForceConflict, "ssa-force-conflicts", false, "If true, server-side apply will force the changes against conflicts.")
}

// ForceFlag returns value to be used in PatchOpts depending ona patch type and SSA mode
func (s SSAFlags) ForceParamValue(patchType types.PatchType) *bool {
	if patchType == types.ApplyPatchType && s.SSAEnable {
		var t = true
		return &t
	} else {
		// nil cats like False for ApplyPatchType
		return nil
	}
}
