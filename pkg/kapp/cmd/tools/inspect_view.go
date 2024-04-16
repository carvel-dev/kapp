// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"fmt"

	ctlcap "carvel.dev/kapp/pkg/kapp/clusterapply"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type InspectView struct {
	Source    string
	Resources []ctlres.Resource
	Sort      bool
}

func (v InspectView) Print(ui ui.UI) {
	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	reconcileStateHeader := uitable.NewHeader("Reconcile state")
	reconcileStateHeader.Title = "Rs"

	reconcileInfoHeader := uitable.NewHeader("Reconcile info")
	reconcileInfoHeader.Title = "Ri"

	table := uitable.Table{
		Title:   fmt.Sprintf("Resources in %s", v.Source),
		Content: "resources",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Kind"),
			versionHeader,
			uitable.NewHeader("Owner"),
			reconcileStateHeader,
			reconcileInfoHeader,
			uitable.NewHeader("Age"),
		},

		Notes: []string{"Rs: Reconcile state", "Ri: Reconcile information"},
	}

	if v.Sort {
		table.SortBy = []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true},
			{Column: 2, Asc: true},
			{Column: 3, Asc: true},
		}
	} else {
		// Otherwise it might look very awkward
		table.FillFirstColumn = true
	}

	for _, resource := range v.Resources {
		row := []uitable.Value{
			cmdcore.NewValueNamespace(resource.Namespace()),
			uitable.NewValueString(resource.Name()),
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
			NewValueResourceOwner(resource),
		}

		if resource.IsProvisioned() {
			syncVal := ctlcap.NewValueResourceConverged(resource)

			row = append(row,
				syncVal.StateVal,
				syncVal.ReasonVal,
				cmdcore.NewValueAge(resource.CreatedAt()),
			)
		} else {
			row = append(row,
				uitable.NewValueString(""),
				uitable.NewValueString(""),
				uitable.NewValueString(""),
			)
		}

		table.Rows = append(table.Rows, row)
	}

	ui.PrintTable(table)
}

func NewValueResourceOwner(resource ctlres.Resource) uitable.ValueString {
	if resource.IsProvisioned() {
		if resource.Transient() {
			return uitable.NewValueString("cluster")
		}
		return uitable.NewValueString("kapp")
	}
	return uitable.NewValueString("")
}
