// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
)

type UIFlags struct {
	TTY            bool
	Color          bool
	JSON           bool
	NonInteractive bool
	Columns        []string
}

func (f *UIFlags) Set(cmd *cobra.Command, _ cmdcore.FlagsFactory) {
	cmd.PersistentFlags().BoolVar(&f.Color, "color", true, "Set color output")
	cmd.PersistentFlags().BoolVar(&f.JSON, "json", false, "Output as JSON")
	cmd.PersistentFlags().BoolVarP(&f.NonInteractive, "yes", "y", false, "Assume yes for any prompt")
	cmd.PersistentFlags().StringSliceVar(&f.Columns, "column", nil, "Filter to show only given columns")
}

func (f *UIFlags) ConfigureUI(ui *ui.ConfUI) {
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
