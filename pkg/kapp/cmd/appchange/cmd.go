// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-change",
		Aliases: []string{"ac", "app-changes", "appchange", "appchanges"},
		Short:   "App change",
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	return cmd
}
