// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type WarningFlags struct {
	NoDeprecationWarnings bool
}

func (f *WarningFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	cmd.PersistentFlags().BoolVar(&f.NoDeprecationWarnings, "no-deprecation-warnings", false, "Silence deprecation warnings")
}

func (f *WarningFlags) Configure(depsFactory cmdcore.DepsFactory) {
	depsFactory.ConfigureDeprecationWarning(f.NoDeprecationWarnings)
}
