// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type WarningFlags struct {
	Warnings bool
}

func (f *WarningFlags) Set(cmd *cobra.Command, _ cmdcore.FlagsFactory) {
	cmd.PersistentFlags().BoolVar(&f.Warnings, "warnings", true, "Show warnings")
}

func (f *WarningFlags) Configure(depsFactory cmdcore.DepsFactory) {
	depsFactory.ConfigureWarnings(f.Warnings)
}
