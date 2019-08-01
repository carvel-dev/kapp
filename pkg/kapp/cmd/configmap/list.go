package configmap

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
		Short:   "List config maps",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	cmd.Flags().BoolVar(&o.Values, "values", false, "Show config map values")
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

	app, err := ctlapp.NewApps(o.AppFlags.NamespaceFlags.Name, coreClient, dynamicClient).Find(o.AppFlags.Name)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	identifiedResources := ctlres.NewIdentifiedResources(coreClient, dynamicClient, []string{o.AppFlags.NamespaceFlags.Name})

	maps, err := identifiedResources.ConfigMapResources(labelSelector)
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
				uitable.NewValueString(m.Namespace),
				uitable.NewValueString(m.Name),
				uitable.NewValueString(""),
				uitable.NewValueString(""),
			}

			table.Rows = append(table.Rows, row)
		}

		for k, v := range m.Data {
			row := []uitable.Value{
				uitable.NewValueString(m.Namespace),
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
