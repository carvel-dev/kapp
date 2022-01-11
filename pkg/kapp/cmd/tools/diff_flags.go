// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"fmt"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	"github.com/k14s/kapp/pkg/kapp/cmd/tools/ssa"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/spf13/cobra"
)

type DiffFlags struct {
	ctlcap.ChangeSetViewOpts
	ctldiff.ChangeSetOpts
	ctldiff.ChangeSetFilter

	Run        bool
	ExitStatus bool
	UI         bool
}

func (s *DiffFlags) SetWithPrefix(prefix string, cmd *cobra.Command) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.Run, prefix+"run", false, "Show diff and exit successfully without any further action")
	cmd.Flags().BoolVar(&s.ExitStatus, prefix+"exit-status", false, "Return specific exit status based on number of changes")
	cmd.Flags().BoolVar(&s.UI, prefix+"ui-alpha", false, "Start UI server to inspect changes (alpha feature)")

	cmd.Flags().BoolVar(&s.Summary, prefix+"summary", true, "Show diff summary")
	cmd.Flags().BoolVarP(&s.Changes, prefix+"changes", "c", false, "Show changes")

	cmd.Flags().IntVar(&s.Context, prefix+"context", 2, "Show number of lines around changed lines")
	cmd.Flags().BoolVar(&s.LineNumbers, prefix+"line-numbers", true, "Show line numbers")
	cmd.Flags().BoolVar(&s.Mask, prefix+"mask", true, "Apply masking rules")

	if *cmd.Flags().Bool(prefix+"against-last-applied", true, "Show changes against last applied copy when possible. (Conflicts with server-side apply)") {
		s.ChangeSetOpts.Mode = ctldiff.AgainstLastAppliedChangeSetMode
	} else {
		s.ChangeSetOpts.Mode = ctldiff.ExactChangeSetMode
	}

	cmd.Flags().StringVar(&s.Filter, prefix+"filter", "", `Set changes filter (example: {"and":[{"ops":["update"]},{"existingResource":{"kinds":["Deployment"]}]})`)
}

func AdjustDiffFlags(ssa ssa.SSAFlags, df *DiffFlags, diffPrefix string, cmd *cobra.Command) error {
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
