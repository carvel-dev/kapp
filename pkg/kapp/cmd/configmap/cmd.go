// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package configmap

import (
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config-map",
		Aliases: []string{"cm", "cfg", "config-maps", "configmap", "configmaps"},
		Short:   "Config map",
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	return cmd
}
