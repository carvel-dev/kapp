// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"github.com/spf13/cobra"
)

type LoggerFlags struct {
	Debug bool
}

func (f *LoggerFlags) Set(cmd *cobra.Command, _ cmdcore.FlagsFactory) {
	cmd.PersistentFlags().BoolVar(&f.Debug, "debug", false, "Include debug output")
}

func (f *LoggerFlags) Configure(logger *logger.UILogger) {
	logger.SetDebug(f.Debug)
}
