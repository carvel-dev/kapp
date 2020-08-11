// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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
