// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
)

type Flags struct {
	NamespaceFlags cmdcore.NamespaceFlags
	Name           string
	AppNamespace   string
}

func (s *Flags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.Name, "group", "g", "", "Set app group name")
	cmd.Flags().StringVar(&s.AppNamespace, "app-namespace", s.AppNamespace, "Set app namespace (to store app state)")
}
