// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"
	"strings"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/mitchellh/go-wordwrap"
)

type ChangeView interface {
	Resource() ctlres.Resource
	ClusterOriginalResource() ctlres.Resource

	ApplyOp() ClusterChangeApplyOp
	ApplyStrategyOp() (ClusterChangeApplyStrategyOp, error)
	WaitOp() ClusterChangeWaitOp
	ConfigurableTextDiff() *ctldiff.ConfigurableTextDiff
}

type ChangesView struct {
	ChangeViews []ChangeView
	Sort        bool

	countsView *ChangesCountsView
}

func (v *ChangesView) Print(ui ui.UI) {
	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	reconcileStateHeader := uitable.NewHeader("Reconcile state")
	reconcileStateHeader.Title = "Rs"

	reconcileInfoHeader := uitable.NewHeader("Reconcile info")
	reconcileInfoHeader.Title = "Ri"

	opStrategyHeader := uitable.NewHeader("Op strategy")
	opStrategyHeader.Title = "Op st."

	table := uitable.Table{
		Title: "Changes",
		// TODO do not show total number of "changes" as it may
		// be confusing that some changes are only waits
		// Content: "changes",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Kind"),
			versionHeader,
			uitable.NewHeader("Age"),
			uitable.NewHeader("Op"),
			opStrategyHeader,
			uitable.NewHeader("Wait to"),
			reconcileStateHeader,
			reconcileInfoHeader,
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

	for _, view := range v.ChangeViews {
		resource := view.Resource()
		v.countsView.Add(view.ApplyOp(), view.WaitOp())

		row := []uitable.Value{
			cmdcore.NewValueNamespace(resource.Namespace()),
			uitable.NewValueString(resource.Name()),
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
		}

		if resource.IsProvisioned() {

			row = append(row,
				cmdcore.NewValueAge(resource.CreatedAt()),
			)
		} else {
			row = append(row,
				uitable.NewValueString(""),
			)
		}

		row = append(row,
			v.applyOpCode(view.ApplyOp()),
			v.applyStrategyOpCode(view),
			v.waitOpCode(view.WaitOp()),
		)

		if view.ClusterOriginalResource() != nil {
			syncVal := NewValueResourceConverged(view.ClusterOriginalResource())
			row = append(row, syncVal.StateVal, syncVal.ReasonVal)
		} else {
			row = append(row,
				uitable.NewValueString(""),
				uitable.NewValueString(""),
			)
		}

		table.Rows = append(table.Rows, row)
	}

	table.Notes = append(table.Notes, v.countsView.Strings(true)...)

	ui.PrintTable(table)
}

func (v *ChangesView) Summary() string { return v.countsView.String() }

var (
	applyOpCodeUI = map[ClusterChangeApplyOp]string{
		ClusterChangeApplyOpAdd:    "create",
		ClusterChangeApplyOpDelete: "delete",
		ClusterChangeApplyOpUpdate: "update",
		ClusterChangeApplyOpNoop:   "noop",
		ClusterChangeApplyOpExists: "exists",
	}

	applyStrategyCodeUI = map[ClusterChangeApplyOp]map[ClusterChangeApplyStrategyOp]string{
		ClusterChangeApplyOpAdd: {
			createStrategyPlainAnnValue:                  "",
			createStrategyFallbackOnUpdateAnnValue:       "fallback on update",
			createStrategyFallbackOnUpdateOrNoopAnnValue: "fallback on update or noop",
		},

		ClusterChangeApplyOpUpdate: {
			updateStrategyPlainAnnValue:             "",
			updateStrategyFallbackOnReplaceAnnValue: "fallback on replace",
			updateStrategyAlwaysReplaceAnnValue:     "always replace",
			updateStrategySkipAnnValue:              "skip",
		},

		ClusterChangeApplyOpDelete: {
			deleteStrategyPlainAnnValue:  "",
			deleteStrategyOrphanAnnValue: "orphan",
		},

		ClusterChangeApplyOpNoop: {
			noopStrategyOp: "",
		},

		ClusterChangeApplyOpExists: {
			noopStrategyOp: "",
		},
	}

	waitOpCodeUI = map[ClusterChangeWaitOp]string{
		ClusterChangeWaitOpOK:     "reconcile",
		ClusterChangeWaitOpDelete: "delete",
		ClusterChangeWaitOpNoop:   "noop",
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
	case ClusterChangeApplyOpExists:
		return uitable.ValueFmt{V: uitable.NewValueString(applyOpCodeUI[op]), Error: false}
	case ClusterChangeApplyOpNoop:
		return uitable.NewValueString("")
	default:
		return uitable.NewValueString("???")
	}
}

func (v *ChangesView) applyStrategyOpCode(view ChangeView) uitable.Value {
	strategyOp, err := view.ApplyStrategyOp()
	if err == nil {
		if codeUIs, found := applyStrategyCodeUI[view.ApplyOp()]; found {
			if codeUI, found := codeUIs[strategyOp]; found {
				return uitable.NewValueString(codeUI)
			}
		}
	}
	return uitable.NewValueString("???")
}

func (v *ChangesView) waitOpCode(op ClusterChangeWaitOp) uitable.Value {
	switch op {
	case ClusterChangeWaitOpOK:
		return uitable.NewValueString(waitOpCodeUI[op]) // TODO highlight for apply op noop?
	case ClusterChangeWaitOpDelete:
		return uitable.NewValueString(waitOpCodeUI[op])
	case ClusterChangeWaitOpNoop:
		return uitable.NewValueString("")
	default:
		return uitable.NewValueString("???")
	}
}

type ChangesCountsView struct {
	applyOps map[ClusterChangeApplyOp]int
	waitOps  map[ClusterChangeWaitOp]int
}

func NewChangesCountsView() *ChangesCountsView {
	return &ChangesCountsView{map[ClusterChangeApplyOp]int{}, map[ClusterChangeWaitOp]int{}}
}

func (v *ChangesCountsView) Add(applyOp ClusterChangeApplyOp, waitOp ClusterChangeWaitOp) {
	v.applyOps[applyOp]++
	v.waitOps[waitOp]++
}

func (v *ChangesCountsView) Strings(table bool) []string {
	applyOpsStats := []string{}
	visibleApplyOps := []ClusterChangeApplyOp{
		ClusterChangeApplyOpAdd, ClusterChangeApplyOpDelete, ClusterChangeApplyOpUpdate, ClusterChangeApplyOpNoop, ClusterChangeApplyOpExists}

	for _, op := range visibleApplyOps {
		applyOpsStats = append(applyOpsStats, fmt.Sprintf("%d %s", v.applyOps[op], applyOpCodeUI[op]))
	}

	waitsOpStats := []string{}
	visibleWaitOps := []ClusterChangeWaitOp{
		ClusterChangeWaitOpOK, ClusterChangeWaitOpDelete, ClusterChangeWaitOpNoop}

	for _, op := range visibleWaitOps {
		waitsOpStats = append(waitsOpStats, fmt.Sprintf("%d %s", v.waitOps[op], waitOpCodeUI[op]))
	}

	padding := ""
	if table {
		padding = "     "
	}

	return []string{
		"Op: " + padding + strings.Join(applyOpsStats, ", "),
		"Wait to: " + strings.Join(waitsOpStats, ", "),
	}
}

func (v *ChangesCountsView) String() string {
	return strings.Join(v.Strings(false), " / ")
}

type ValueResourceConverged struct {
	StateVal  uitable.Value
	ReasonVal uitable.Value
}

func NewValueResourceConverged(resource ctlres.Resource) ValueResourceConverged {
	// TODO how to retrieve waiting rules
	convergedResFactory := NewConvergedResourceFactory(nil, ConvergedResourceFactoryOpts{})

	// TODO state vs err vs output
	state, _, err := convergedResFactory.New(resource, nil).IsDoneApplying()
	stateUI := NewDoneApplyStateUI(state, err)

	stateVal := uitable.ValueFmt{V: uitable.NewValueString(stateUI.State), Error: stateUI.Error}
	reasonVal := uitable.NewValueString(wordwrap.WrapString(stateUI.Message, 35))

	return ValueResourceConverged{stateVal, reasonVal}
}
