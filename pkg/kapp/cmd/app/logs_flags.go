package app

import (
	"fmt"

	ctllogs "github.com/k14s/kapp/pkg/kapp/logs"
	"github.com/spf13/cobra"
)

type LogsFlags struct {
	Follow       bool
	Lines        int64
	ContainerTag bool
	PodName      string
}

func (s *LogsFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&s.Follow, "follow", "f", false, "As new pods are added, new pod logs will be printed")
	cmd.Flags().Int64Var(&s.Lines, "lines", 10, "Number of lines")
	cmd.Flags().BoolVar(&s.ContainerTag, "container-tag", true, "Include container tag")
	cmd.Flags().StringVarP(&s.PodName, "pod-name", "m", "", "Set partial pod name to filter logs")
}

func (s *LogsFlags) PodLogOpts() (ctllogs.PodLogOpts, error) {
	if !s.Follow && s.Lines <= 0 {
		return ctllogs.PodLogOpts{}, fmt.Errorf(
			"Expected --lines to be greater than zero since --follow is not specified")
	}

	opts := ctllogs.PodLogOpts{Follow: s.Follow, ContainerTag: s.ContainerTag}

	if !s.Follow {
		opts.Lines = &s.Lines
	}

	return opts, nil
}
