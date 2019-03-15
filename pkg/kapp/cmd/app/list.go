package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	NamespaceFlags cmdcore.NamespaceFlags
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List all apps in a namespace",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.NamespaceFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ListOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	dynamicClient, err := o.depsFactory.DynamicClient()
	if err != nil {
		return err
	}

	apps := ctlapp.NewApps(o.NamespaceFlags.Name, coreClient, dynamicClient)

	items, err := apps.List(nil)
	if err != nil {
		return err
	}

	table := uitable.Table{
		Title:   fmt.Sprintf("Apps in namespace '%s'", o.NamespaceFlags.Name),
		Content: "apps",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Label"),
			uitable.NewHeader("Namespaces"),
			uitable.NewHeader("Last Change Successful"),
			uitable.NewHeader("Last Change Age"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	for _, item := range items {
		sel, err := item.LabelSelector()
		if err != nil {
			return err
		}

		row := []uitable.Value{
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
