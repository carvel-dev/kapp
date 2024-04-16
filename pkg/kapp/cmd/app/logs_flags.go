// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	ctllogs "carvel.dev/kapp/pkg/kapp/logs"
	"github.com/spf13/cobra"
)

type LogsFlags struct {
	Follow         bool
	Lines          int64
	ContainerNames []string
	ContainerTag   bool
	PodName        string
}

func (s *LogsFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&s.Follow, "follow", "f", false, "As new pods are added, new pod logs will be printed")
	cmd.Flags().Int64Var(&s.Lines, "lines", 10, "Limit to number of lines (use -1 to remove limit)")

	cmd.Flags().BoolVar(&s.ContainerTag, "container-tag", true, "Include container tag")

	cmd.Flags().StringVarP(&s.PodName, "pod-name", "m", "",
		"Set pod name to filter logs (% acts as wildcard, e.g. 'app%')")

	cmd.Flags().StringSliceVarP(&s.ContainerNames, "container-name", "c", nil,
		"Set container name to filter logs (% acts as wildcard, e.g. 'app%') (can repeat)")
}

func (s *LogsFlags) PodLogOpts() (ctllogs.PodLogOpts, error) {
	if !s.Follow && s.Lines <= 0 {
		return ctllogs.PodLogOpts{}, fmt.Errorf(
			"Expected --lines to be greater than zero since --follow is not specified")
	}

	opts := ctllogs.PodLogOpts{Follow: s.Follow, ContainerNames: s.ContainerNames, ContainerTag: s.ContainerTag}

	if s.Lines >= 0 {
		opts.Lines = &s.Lines
	}

	return opts, nil
}
