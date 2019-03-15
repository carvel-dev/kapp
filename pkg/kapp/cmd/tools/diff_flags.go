package tools

import (
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/spf13/cobra"
)

type DiffFlags struct {
	ctldiff.ChangeSetViewOpts
	ctldiff.ChangeSetOpts

	Run bool
}

func (s *DiffFlags) SetWithPrefix(prefix string, cmd *cobra.Command) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.Run, prefix+"run", false, "Show diff and exit successfully without any further action")

	cmd.Flags().BoolVar(&s.Summary, prefix+"summary", true, "Show diff summary")
	cmd.Flags().BoolVar(&s.SummaryFull, prefix+"summary-full", false, "Show full diff summary (includes ignored and unchanged items)")

	cmd.Flags().BoolVar(&s.Changes, prefix+"changes", false, "Show changes")
	cmd.Flags().BoolVar(&s.ChangesFull, prefix+"changes-full", false, "Show full changes (includes ignored and unchanged items)")

	cmd.Flags().IntVar(&s.Context, prefix+"context", 2, "Show number of lines around changed lines")

	cmd.Flags().BoolVar(&s.AgainstLastApplied, prefix+"against-last-applied", true, "Show changes against last applied copy when possible (if set to false, always use live copy from the server)")
}
