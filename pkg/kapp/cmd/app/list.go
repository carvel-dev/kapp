package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	NamespaceFlags cmdcore.NamespaceFlags
	AllNamespaces  bool
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List all apps in a namespace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppHelpGroup.Key: cmdcore.AppHelpGroup.Value,
		},
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false, "List apps in all namespaces")
	return cmd
}

func (o *ListOptions) Run() error {
	tableTitle := fmt.Sprintf("Apps in namespace '%s'", o.NamespaceFlags.Name)
	nsHeader := uitable.NewHeader("Namespace")
	nsHeader.Hidden = true

	if o.AllNamespaces {
		o.NamespaceFlags.Name = ""
		tableTitle = "Apps in all namespaces"
		nsHeader.Hidden = false
	}

	apps, _, _, err := AppFactoryClients(o.depsFactory, o.NamespaceFlags, o.logger)
	if err != nil {
		return err
	}

	items, err := apps.List(nil)
	if err != nil {
		return err
	}

	lcsHeader := uitable.NewHeader("Last Change Successful")
	lcsHeader.Title = "Lcs"

	lcaHeader := uitable.NewHeader("Last Change Age")
	lcaHeader.Title = "Lca"

	table := uitable.Table{
		Title:   tableTitle,
		Content: "apps",

		Header: []uitable.Header{
			nsHeader,
			uitable.NewHeader("Name"),
			uitable.NewHeader("Label"),
			uitable.NewHeader("Namespaces"),
			lcsHeader,
			lcaHeader,
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true},
		},

		Notes: []string{
			lcsHeader.Title + ": Last Change Successful",
			lcaHeader.Title + ": Last Change Age",
		},
	}

	for _, item := range items {
		sel, err := item.LabelSelector()
		if err != nil {
			return err
		}

		row := []uitable.Value{
			cmdcore.NewValueNamespace(item.Namespace()),
			uitable.NewValueString(item.Name()),
			uitable.NewValueString(sel.String()),
		}

		lastChange, err := item.LastChange()
		if err != nil {
			return err
		}

		if lastChange != nil {
			row = append(row,
				uitable.NewValueString(strings.Join(lastChange.Meta().Namespaces, ",")),
				uitable.ValueFmt{
					V:     cmdcore.NewValueUnknownBool(lastChange.Meta().Successful),
					Error: lastChange.Meta().Successful == nil || *lastChange.Meta().Successful != true,
				},
				cmdcore.NewValueAge(lastChange.Meta().StartedAt),
			)
		} else {
			row = append(row,
				uitable.NewValueString(""),
				cmdcore.NewValueUnknownBool(nil),
				cmdcore.NewValueAge(time.Time{}),
			)
		}

		table.Rows = append(table.Rows, row)
	}

	o.ui.PrintTable(table)

	return nil
}
