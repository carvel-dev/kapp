package diff

import (
	"fmt"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
)

type ChangesView struct {
	Changes     []Change
	Sort        bool
	IncludeFunc func(ch Change) bool

	summary string
}

func (v *ChangesView) Print(ui ui.UI) {
	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	ignoredHeader := uitable.NewHeader("Ignored")
	ignoredHeader.Hidden = true

	table := uitable.Table{
		Title:   "Changes",
		Content: "changes",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Kind"),
			versionHeader,
			uitable.NewHeader("Conditions"),
			uitable.NewHeader("Age"),
			uitable.NewHeader("Changed"),
			ignoredHeader,
			uitable.NewHeader("Ignored Reason"),
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

	countsView := NewChangesCountsView()

	for _, change := range v.Changes {
		if !v.IncludeFunc(change) {
			countsView.AddHidden(change.Op())
			continue
		}

		countsView.Add(change.Op())

		resource := change.NewOrExistingResource()

		row := []uitable.Value{
			uitable.NewValueString(resource.Namespace()),
			uitable.NewValueString(resource.Name()),
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
		}

		if resource.IsProvisioned() {
			condVal := cmdcore.NewConditionsValue(resource.Status())

			row = append(row,
				// TODO erroneously colors empty value
				uitable.ValueFmt{V: condVal, Error: condVal.NeedsAttention()},
				cmdcore.NewValueAge(resource.CreatedAt()),
			)
		} else {
			row = append(row,
				uitable.ValueFmt{V: uitable.NewValueString(""), Error: false},
				uitable.NewValueString(""),
			)
		}

		row = append(row,
			v.opCode(change),
			uitable.NewValueBool(change.IsIgnored()),
			uitable.NewValueString(change.IgnoredReason()),
		)

		table.Rows = append(table.Rows, row)
	}

	v.summary = countsView.String()

	table.Notes = []string{v.summary}

	ui.PrintTable(table)
}

func (v *ChangesView) Summary() string { return v.summary }

func (v *ChangesView) opCode(change Change) uitable.Value {
	switch change.Op() {
	case ChangeOpAdd:
		return uitable.ValueFmt{V: uitable.NewValueString("add"), Error: false}
	case ChangeOpDelete:
		return uitable.ValueFmt{V: uitable.NewValueString("del"), Error: true}
	case ChangeOpUpdate:
		return uitable.ValueFmt{V: uitable.NewValueString("mod"), Error: false}
	case ChangeOpKeep:
		return uitable.NewValueString("")
	default:
		return uitable.NewValueString("???")
	} // TODO yellow color?
}

type ChangesCountsView struct {
	all    map[ChangeOp]int
	hidden map[ChangeOp]int
}

func NewChangesCountsView() *ChangesCountsView {
	return &ChangesCountsView{map[ChangeOp]int{}, map[ChangeOp]int{}}
}

func (v *ChangesCountsView) Add(op ChangeOp)       { v.all[op] += 1 }
func (v *ChangesCountsView) AddHidden(op ChangeOp) { v.hidden[op] += 1 }

func (v *ChangesCountsView) String() string {
	result := []string{}
	for _, op := range allChangeOps {
		hiddenStr := ""
		if v.hidden[op] > 0 {
			hiddenStr = fmt.Sprintf(" (%d hidden)", v.hidden[op])
		}
		result = append(result, fmt.Sprintf("%d %s%s", v.all[op], op, hiddenStr))
	}
	return strings.Join(result, ", ")
}
