package tools

import (
	"github.com/spf13/cobra"
)

type FileFlags struct {
	Files     []string
	Recursive bool
	Sort      bool
}

func (s *FileFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.Files, "file", "f", nil, "Set file (format: /tmp/foo, https://..., -) (can be specified multiple times)")
	cmd.Flags().BoolVarP(&s.Recursive, "recursive", "R", false, "Process directory used in -f recursively")
	cmd.Flags().BoolVar(&s.Sort, "sort", true, "Sort by namespace, name, etc.")
}

type FileFlags2 struct {
	Files     []string
	Recursive bool
}

func (s *FileFlags2) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.Files, "file2", nil, "Set second file (format: /tmp/foo, https://..., -) (can be specified multiple times)")
	cmd.Flags().BoolVar(&s.Recursive, "file2-recursive", false, "Process directory used in --file2 recursively")
}
