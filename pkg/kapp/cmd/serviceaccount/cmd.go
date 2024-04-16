// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package serviceaccount

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-account",
		Aliases: []string{"sa", "service-accounts", "serviceccounts", "serviceaccount"},
		Short:   "Service account",
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	return cmd
}
