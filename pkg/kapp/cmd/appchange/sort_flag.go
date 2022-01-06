package appchange

import "github.com/spf13/cobra"

type SortFlag struct {
	Sort string
}

func (s *SortFlag) Set(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.Sort, "sort", "s", "newest-first", "Sort app-changes list (allowed values : newest-first, oldest-first)")
}

func (s *SortFlag) IsSortByNewestFirst() bool {
	if s.Sort == "newest-first" {
		return true
	}
	return false
}
