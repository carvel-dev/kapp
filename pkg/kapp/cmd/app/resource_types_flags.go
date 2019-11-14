package app

import (
	"github.com/spf13/cobra"
)

type ResourceTypesFlags struct {
	IgnoreFailingAPIServices bool
}

func (s *ResourceTypesFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.IgnoreFailingAPIServices, "dangerous-ignore-failing-api-services",
		false, "Allow to ignore failing APIServices")
}
