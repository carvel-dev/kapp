package clusterapply

import (
	"fmt"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeView interface {
	Resource() ctlres.Resource
	ApplyOp() ClusterChangeApplyOp
	WaitOp() ClusterChangeWaitOp
	TextDiff() ctldiff.TextDiff
	IsIgnored() bool
	IgnoredReason() string
}

type ChangesView struct {
	ChangeViews []ChangeView
	Sort        bool

	summary string
}

func (v *ChangesView) Print(ui ui.UI) {
	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	conditionsHeader := uitable.NewHeader("Conditions")
	conditionsHeader.Title = "Conds."

	table := uitable.Table{
		Title:   "Changes",
		Content: "changes",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Kind"),
			versionHeader,
			conditionsHeader,
			uitable.NewHeader("Age"),
			uitable.NewHeader("Op"),
			uitable.NewHeader("Wait to"),
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

	for _, view := range v.ChangeViews {
		resource := view.Resource()
		countsView.Add(view.ApplyOp())

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
			v.applyOpCode(view.ApplyOp()),
			v.waitOpCode(view.WaitOp()),
		)

		table.Rows = append(table.Rows, row)
	}

	v.summary = countsView.String()

	table.Notes = []string{v.summary}

	ui.PrintTable(table)
}

func (v *ChangesView) Summary() string { return v.summary }

var (
	applyOpCodeUI = map[ClusterChangeApplyOp]string{
		ClusterChangeApplyOpAdd:    "create",
		ClusterChangeApplyOpDelete: "delete",
		ClusterChangeApplyOpUpdate: "update",
		ClusterChangeApplyOpNoop:   "",
	}

	waitOpCodeUI = map[ClusterChangeWaitOp]string{
		ClusterChangeWaitOpOK:     "reconcile",
		ClusterChangeWaitOpDelete: "delete",
		ClusterChangeWaitOpNoop:   "",
	}
)

func (v *ChangesView) applyOpCode(op ClusterChangeApplyOp) uitable.Value {
	switch op {
	case ClusterChangeApplyOpAdd:
		return uitable.ValueFmt{V: uitable.NewValueString(applyOpCodeUI[op]), Error: false}
	case ClusterChangeApplyOpDelete:
		return uitable.ValueFmt{V: uitable.NewValueString(applyOpCodeUI[op]), Error: true}
	case ClusterChangeApplyOpUpdate:
		return uitable.ValueFmt{V: uitable.NewValueString(applyOpCodeUI[op]), Error: false}
	case ClusterChangeApplyOpNoop:
		return uitable.NewValueString(applyOpCodeUI[op])
	default:
		return uitable.NewValueString("???")
	}
}

func (v *ChangesView) waitOpCode(op ClusterChangeWaitOp) uitable.Value {
	switch op {
	case ClusterChangeWaitOpOK:
		return uitable.NewValueString(waitOpCodeUI[op]) // TODO highlight for apply op noop?
	case ClusterChangeWaitOpDelete:
		return uitable.NewValueString(waitOpCodeUI[op])
	case ClusterChangeWaitOpNoop:
		return uitable.NewValueString(waitOpCodeUI[op])
	default:
		return uitable.NewValueString("???")
	}
}

type ChangesCountsView struct {
	all map[ClusterChangeApplyOp]int
}

func NewChangesCountsView() *ChangesCountsView {
	return &ChangesCountsView{map[ClusterChangeApplyOp]int{}}
}

func (v *ChangesCountsView) Add(op ClusterChangeApplyOp) { v.all[op] += 1 }

func (v *ChangesCountsView) String() string {
	result := []string{}
	for _, op := range allClusterChangeApplyOps {
		result = append(result, fmt.Sprintf("%d %s", v.all[op], applyOpCodeUI[op]))
	}
	return strings.Join(result, ", ")
}
