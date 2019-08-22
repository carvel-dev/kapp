package app

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type InspectStatusView struct {
	Source    string
	Resources []ctlres.Resource
}

func (v InspectStatusView) Print(ui ui.UI) {
	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	table := uitable.Table{
		Title:   fmt.Sprintf("Resources in %s", v.Source),
		Content: "resources",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Kind"),
			versionHeader,
			uitable.NewHeader("Status"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true},
			{Column: 2, Asc: true},
			{Column: 3, Asc: true},
		},

		FillFirstColumn: true, // because of transpose
		Transpose:       true,
	}

	for _, resource := range v.Resources {
		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(resource.Namespace()),
			uitable.NewValueString(resource.Name()),
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
			uitable.NewValueInterface(resource.Status()),
		})
	}

	ui.PrintTable(table)
}
