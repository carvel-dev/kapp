// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/cobra"

	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
)

type Flags struct {
	NamespaceFlags cmdcore.NamespaceFlags
	Name           string
}

func (s *Flags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	s.NamespaceFlags.Set(cmd, flagsFactory)

	cmd.Flags().StringVarP(&s.Name, "app", "a", s.Name, "Set app name (or label selector) (format: name, label:key=val, !key)")
}
