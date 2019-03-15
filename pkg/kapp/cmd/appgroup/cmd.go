package appgroup

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-group",
		Aliases: []string{"ag", "appgroup"},
		Short:   "App group",
	}
	return cmd
}
