package serviceaccount

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-account",
		Aliases: []string{"sa", "service-accounts", "serviceccounts", "serviceaccount"},
		Short:   "Service account",
	}
	return cmd
}
