// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
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
