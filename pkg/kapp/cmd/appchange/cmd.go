package appchange

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-change",
		Aliases: []string{"ac", "app-changes", "appchange", "appchanges"},
		Short:   "App change",
	}
	return cmd
}
