package appgroup

import (
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type AppGroupFlags struct {
	NamespaceFlags cmdcore.NamespaceFlags
	Name           string
}

func (s *AppGroupFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.Name, "group", "g", "", "Set app group name")
}
