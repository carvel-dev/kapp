// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	ctlcap "carvel.dev/kapp/pkg/kapp/clusterapply"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	"github.com/spf13/cobra"
)

type DiffFlags struct {
	ctlcap.ChangeSetViewOpts
	ctldiff.ChangeSetOpts
	ctldiff.ChangeSetFilter

	Run        bool
	ExitStatus bool
	UI         bool

	AnchoredDiff bool
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

	cmd.Flags().BoolVar(&s.AgainstLastApplied, prefix+"against-last-applied", true, "Show changes against last applied copy when possible")

	cmd.Flags().StringVar(&s.Filter, prefix+"filter", "", `Set changes filter (example: {"and":[{"ops":["update"]},{"existingResource":{"kinds":["Deployment"]}]})`)
	cmd.Flags().BoolVar(&s.ChangesYAML, prefix+"changes-yaml", false, "Print YAML to be applied")

	cmd.Flags().BoolVar(&s.AnchoredDiff, prefix+"anchored", false, "Allow using anchored diff for large resources")
}
