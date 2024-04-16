// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
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
			cmdcore.NewValueNamespace(resource.Namespace()),
			uitable.NewValueString(resource.Name()),
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
			uitable.NewValueInterface(resource.Status()),
		})
	}

	ui.PrintTable(table)
}
