package tools

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type InspectView struct {
	Source    string
	Resources []ctlres.Resource
	Sort      bool
}

func (v InspectView) Print(ui ui.UI) {
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
			uitable.NewHeader("Managed by"),
			uitable.NewHeader("Conditions"),
			uitable.NewHeader("Sync\nstate"),
			uitable.NewHeader("Sync\nmsg"),
			uitable.NewHeader("Age"),
		},
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
			uitable.NewValueString(resource.Namespace()),
			uitable.NewValueString(resource.Name()),
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
			NewValueResourceManagedBy(resource),
		}

		if resource.IsProvisioned() {
			condVal := cmdcore.NewConditionsValue(resource.Status())
			syncVal := NewValueResourceConverged(resource)

			row = append(row,
				// TODO erroneously colors empty value
				uitable.ValueFmt{V: condVal, Error: condVal.NeedsAttention()},
				syncVal.StateVal,
				syncVal.ReasonVal,
				cmdcore.NewValueAge(resource.CreatedAt()),
			)
		} else {
			row = append(row,
				uitable.ValueFmt{V: uitable.NewValueString(""), Error: false},
				uitable.NewValueString(""),
				uitable.NewValueString(""),
				uitable.NewValueString(""),
			)
		}

		table.Rows = append(table.Rows, row)
	}

	ui.PrintTable(table)
}

type ValueResourceConverged struct {
	StateVal  uitable.Value
	ReasonVal uitable.Value
}

func NewValueResourceConverged(resource ctlres.Resource) ValueResourceConverged {
	var stateVal, reasonVal uitable.Value

	// TODO state vs err vs output
	state, err := ctlcap.NewConvergedResource(resource, nil).IsDoneApplying(&noopUI{})
	if err != nil {
		stateVal = uitable.ValueFmt{V: uitable.NewValueString("ERR"), Error: true}
		reasonVal = uitable.NewValueString(err.Error())
	} else {
		switch {
		case state.Done && state.Successful:
			stateVal = uitable.ValueFmt{V: uitable.NewValueString("OK"), Error: false}
		case state.Done && !state.Successful:
			stateVal = uitable.ValueFmt{V: uitable.NewValueString("FAIL"), Error: true}
		case !state.Done:
			stateVal = uitable.ValueFmt{V: uitable.NewValueString("In progress"), Error: true}
		}
		reasonVal = uitable.NewValueString(state.Message)
	}

	return ValueResourceConverged{stateVal, reasonVal}
}

type noopUI struct{}

func (b *noopUI) NotifySection(msg string, args ...interface{}) {}
func (b *noopUI) Notify(msg string, args ...interface{})        {}

func NewValueResourceManagedBy(resource ctlres.Resource) uitable.ValueString {
	if resource.IsProvisioned() {
		if resource.Transient() {
			return uitable.NewValueString("cluster")
		}
		return uitable.NewValueString("kapp")
	}
	return uitable.NewValueString("")
}
