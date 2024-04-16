// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/cobra"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
)

type Flags struct {
	NamespaceFlags cmdcore.NamespaceFlags
	Name           string
	AppNamespace   string
}

func (s *Flags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.Name, "app", "a", s.Name, "Set app name (or label selector) (format: name, label:key=val, !key)")
	cmd.Flags().StringVar(&s.AppNamespace, "app-namespace", s.AppNamespace, "Set app namespace (to store app state)")
}
