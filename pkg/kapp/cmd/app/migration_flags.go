// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/cobra"
)

type MigrationFlags struct {
	PrevAppName string
}

func (s *MigrationFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.PrevAppName, "prev-app", "", "Name of old app incase of migration")
}
