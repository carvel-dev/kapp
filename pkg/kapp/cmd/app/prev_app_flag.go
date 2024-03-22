// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/cobra"
)

type PrevAppFlags struct {
	PrevAppName string
}

func (s *PrevAppFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.PrevAppName, "prev-app", "", "Set previous app name")
}
