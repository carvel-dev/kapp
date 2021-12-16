// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"fmt"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/spf13/cobra"
)

type SSAFlags struct {
	SSAEnable        bool
	SSAConflict      bool
	FieldManagerName string
}

func (s *SSAFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.SSAEnable, "ssa-enable", false, "Use server side apply")
	cmd.Flags().StringVar(&s.FieldManagerName, "ssa-field-manager", "kapp-server-side-apply", "Name of the manager used to track field ownership")
	cmd.Flags().BoolVar(&s.SSAConflict, "ssa-force-conflicts", false, "If true, server-side apply will force the changes against conflicts.")
}

func AdjustDiffFlags(ssa SSAFlags, df *DiffFlags, diffPrefix string, cmd *cobra.Command) error {
	if len(diffPrefix) > 0 {
		diffPrefix = diffPrefix + "-"
	}
	if ssa.SSAEnable {
		alaFlagName := diffPrefix + "against-last-applied"
		if cmd.Flag(alaFlagName).Changed {
			return fmt.Errorf("--ssa-enable conflicts with --%s", alaFlagName)
		}
		df.ChangeSetOpts.Mode = ctldiff.ServerSideApplyChangeSetMode
	}
	return nil
}
