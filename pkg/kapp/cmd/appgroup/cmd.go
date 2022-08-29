// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
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
