// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-group",
		Aliases: []string{"ag", "appgroup"},
		Short:   "app-group will deploy/delete an application for each subdirectory within a directory",
		Example: "$ ls my-repo\n.    ..    app1/    app2/    app3/\n\n$ kapp app-group deploy -g my-env --directory my-repo",
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	return cmd
}
