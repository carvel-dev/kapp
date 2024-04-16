// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package configmap

import (
	cmdapp "carvel.dev/kapp/pkg/kapp/cmd/app"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags cmdapp.Flags
	Values   bool
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List config maps",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	cmd.Flags().BoolVar(&o.Values, "values", false, "Show config map values")
	return cmd
}

func (o *ListOptions) Run() error {
	app, supportObjs, err := cmdapp.Factory(o.depsFactory, o.AppFlags, cmdapp.ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	maps, err := supportObjs.IdentifiedResources.ConfigMapResources(labelSelector)
	if err != nil {
		return err
	}

	valueHeader := uitable.NewHeader("Value")
	valueHeader.Hidden = !o.Values

	table := uitable.Table{
		Title:   "Config map items",
		Content: "config map items",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Key"),
			valueHeader,
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true},
			{Column: 2, Asc: true},
		},
	}

	for _, m := range maps {
		// Show config map info even if it's empty
		if len(m.Data) == 0 {
			row := []uitable.Value{
				cmdcore.NewValueNamespace(m.Namespace),
				uitable.NewValueString(m.Name),
				uitable.NewValueString(""),
				uitable.NewValueString(""),
			}

			table.Rows = append(table.Rows, row)
		}

		for k, v := range m.Data {
			row := []uitable.Value{
				cmdcore.NewValueNamespace(m.Namespace),
				uitable.NewValueString(m.Name),
				uitable.NewValueString(k),
				uitable.NewValueString(v),
			}

			table.Rows = append(table.Rows, row)
		}
	}

	o.ui.PrintTable(table)

	return nil
}
