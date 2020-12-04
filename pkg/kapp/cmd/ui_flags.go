// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type UIFlags struct {
	TTY            bool
	Color          bool
	JSON           bool
	NonInteractive bool
	Columns        []string
}

func (f *UIFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	// Default tty to true: https://github.com/vmware-tanzu/carvel-kapp/issues/28
	cmd.PersistentFlags().BoolVar(&f.TTY, "tty", true, "Force TTY-like output")
	cmd.PersistentFlags().BoolVar(&f.Color, "color", true, "Set color output")
	cmd.PersistentFlags().BoolVar(&f.JSON, "json", false, "Output as JSON")
	cmd.PersistentFlags().BoolVarP(&f.NonInteractive, "yes", "y", false, "Assume yes for any prompt")
	cmd.PersistentFlags().StringSliceVar(&f.Columns, "column", nil, "Filter to show only given columns")
}

func (f *UIFlags) ConfigureUI(ui *ui.ConfUI) {
	ui.EnableTTY(f.TTY)

	if f.Color {
		ui.EnableColor()
	}

	if f.JSON {
		ui.EnableJSON()
	}

	if f.NonInteractive {
		ui.EnableNonInteractive()
	}

	if len(f.Columns) > 0 {
		headers := []uitable.Header{}
		for _, col := range f.Columns {
			headers = append(headers, uitable.Header{
				Key:    uitable.KeyifyHeader(col),
				Hidden: false,
			})
		}

		ui.ShowColumns(headers)
	}
}
