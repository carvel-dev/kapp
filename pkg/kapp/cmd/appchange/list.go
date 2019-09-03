package appchange

import (
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags cmdapp.AppFlags
	Values   bool
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List app changes",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ListOptions) Run() error {
	app, _, _, err := cmdapp.AppFactory(o.depsFactory, o.AppFlags)
	if err != nil {
		return err
	}

	changes, err := app.Changes()
	if err != nil {
		return err
	}

	nsHeader := uitable.NewHeader("Namespaces")
	nsHeader.Hidden = true

	table := uitable.Table{
		Title:   "App changes",
		Content: "app changes",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Started At"),
			uitable.NewHeader("Finished At"),
			uitable.NewHeader("Successful"),
			uitable.NewHeader("Description"),
			nsHeader,
		},

		SortBy: []uitable.ColumnSort{
			{Column: 1, Asc: false},
			{Column: 0, Asc: true}, // in case start time are same
		},
	}

	for _, change := range changes {
		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(change.Name()),
			uitable.NewValueTime(change.Meta().StartedAt),
			uitable.NewValueTime(change.Meta().FinishedAt),
			uitable.ValueFmt{
				V:     cmdcore.NewValueUnknownBool(change.Meta().Successful),
				Error: change.Meta().Successful == nil || *change.Meta().Successful != true,
			},
			uitable.NewValueString(change.Meta().Description),
			uitable.NewValueString(strings.Join(change.Meta().Namespaces, ",")),
		})
	}

	o.ui.PrintTable(table)

	return nil
}
