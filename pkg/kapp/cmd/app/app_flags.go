package app

import (
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type AppFlags struct {
	NamespaceFlags cmdcore.NamespaceFlags
	Name           string
}

func (s *AppFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.Name, "app", "a", "", "Set app name (or label selector) (format: name, label:key=val, !key)")
}
