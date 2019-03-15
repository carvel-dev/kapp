package appgroup

import (
	"github.com/spf13/cobra"
)

type DeployFlags struct {
	Directory string
}

func (s *DeployFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.Directory, "directory", "d", "", "Set directory (format: /tmp/foo)")
}
