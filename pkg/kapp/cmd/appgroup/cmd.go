// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-group",
		Aliases: []string{"ag", "appgroup"},
		Short:   "App group",
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	return cmd
}
