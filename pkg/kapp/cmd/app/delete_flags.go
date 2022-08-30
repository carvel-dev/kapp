// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import "github.com/spf13/cobra"

type DeleteFlags struct {
	DisableCheckingResourceDeletion bool
}

func (s *DeleteFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.DisableCheckingResourceDeletion, "dangerous-disable-checking-resource-deletion",
		false, "Skip checking resource deletion when fully deleting app")
}
