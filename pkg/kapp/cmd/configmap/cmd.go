// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package configmap

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
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
