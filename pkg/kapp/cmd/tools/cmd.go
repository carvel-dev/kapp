// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tools",
		Aliases: []string{"t"},
		Short:   "Tools",
		Annotations: map[string]string{
			cmdcore.MiscHelpGroup.Key: cmdcore.MiscHelpGroup.Value,
		},
	}
	return cmd
}
