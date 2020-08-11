/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package tools

import (
	"fmt"
	"io"
	"strings"

	"github.com/cppforlife/color"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type InspectTreeView struct {
	Source    string
	Resources []ctlres.Resource
	Sort      bool
}

func (v InspectTreeView) Print(ui ui.UI) {
	groupHeader := uitable.NewHeader("Group")
	groupHeader.Hidden = true

	versionHeader := uitable.NewHeader("Version")
	versionHeader.Hidden = true

	conditionsHeader := uitable.NewHeader("Conditions")
	conditionsHeader.Title = "Conds."

	reconcileStateHeader := uitable.NewHeader("Reconcile state")
	reconcileStateHeader.Title = "Rs"

	reconcileInfoHeader := uitable.NewHeader("Reconcile info")
	reconcileInfoHeader.Title = "Ri"

	table := uitable.Table{
		Title:   fmt.Sprintf("Resources in %s", v.Source),
		Content: "resources",

		Header: []uitable.Header{
			groupHeader,
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Kind"),
			versionHeader,
			uitable.NewHeader("Owner"),
			conditionsHeader,
			reconcileStateHeader,
			reconcileInfoHeader,
			uitable.NewHeader("Age"),
		},

		FillFirstColumn: true,

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 5, Asc: true},
			{Column: 1, Asc: true},
		},

		Notes: []string{"Rs: Reconcile state", "Ri: Reconcile information"},
	}

	rsByUID := map[string]ctlres.Resource{}

	for _, resource := range v.Resources {
		rsByUID[resource.UID()] = resource
	}

	for _, resource := range v.Resources {
		prefix := ""
		assocSortingVal := newAssocSortingValue(resource, rsByUID)

		if assocSortingVal.Depth() > 0 {
			prefix = " L" + strings.Repeat("..", assocSortingVal.Depth()-1) + " "
		}

		row := []uitable.Value{
			uitable.NewValueString(assocSortingVal.Value()),
			cmdcore.NewValueNamespace(resource.Namespace()),
			ValueColored{
				// TODO better composability
				S: prefix + resource.Name(),
				Func: func(str string, opts ...interface{}) string {
					result := fmt.Sprintf(str, opts...)
					return strings.Replace(result, prefix, color.New(color.Faint).Sprintf("%s", prefix), 1)
				},
			},
			uitable.NewValueString(resource.Kind()),
			uitable.NewValueString(resource.APIVersion()),
			NewValueResourceOwner(resource),
		}

		if resource.IsProvisioned() {
			condVal := cmdcore.NewConditionsValue(resource.Status())
			syncVal := ctlcap.NewValueResourceConverged(resource)

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

type assocSortingValue struct {
	resource ctlres.Resource
	rsByUID  map[string]ctlres.Resource
	values   []string
}

func newAssocSortingValue(resource ctlres.Resource, rsByUID map[string]ctlres.Resource) *assocSortingValue {
	return &assocSortingValue{resource, rsByUID, nil}
}

func (a *assocSortingValue) Value() string {
	return strings.Join(a.Values(), "@")
}

func (a *assocSortingValue) Values() []string {
	if len(a.values) == 0 {
		a.values = a.calculateValue()
	}
	return a.values
}

func (a *assocSortingValue) calculateValue() []string {
	// TODO currently below prefers label based approach
	// which potentially misses case when some things are
	// labeled and some things are not but both are owner-ref-ed
	lblVal := a.labelAssocStr()
	if len(lblVal) > 0 {
		return []string{lblVal, a.uidOwnersStr()}
	}
	return []string{a.uidOwnersStr()}
}

func (a *assocSortingValue) labelAssocStr() string {
	lblVal := a.resource.Labels()[ctlres.NewAssociationLabel(a.resource).Key()]
	if len(lblVal) > 0 {
		if a.resource.Transient() {
			lblVal = "lbl-" + lblVal + "-2/child" // child
		} else {
			lblVal = "lbl-" + lblVal + "-1" // parent
		}
	}
	return lblVal
}

func (a *assocSortingValue) uidOwnersStr() string {
	uids := []string{a.resource.UID()}
	nextRes := &a.resource

	for nextRes != nil {
		res := *nextRes
		nextRes = nil

		for _, ref := range res.OwnerRefs() {
			foundRes, found := a.rsByUID[string(ref.UID)]
			if found {
				// only nest into first object that we find
				uids = append([]string{string(ref.UID)}, uids...)
				nextRes = &foundRes
				break
			}
		}
	}

	return "ref-" + strings.Join(uids, "/")
}

func (a *assocSortingValue) Depth() int {
	var depth int
	for _, val := range a.Values() {
		depth = a.max(depth, strings.Count(val, "/"))
	}
	return depth
}

func (*assocSortingValue) max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type ValueColored struct {
	S    string
	Func func(string, ...interface{}) string
}

func (t ValueColored) String() string                  { return t.S }
func (t ValueColored) Value() uitable.Value            { return t }
func (t ValueColored) Compare(other uitable.Value) int { panic("Never called") }

func (t ValueColored) Fprintf(w io.Writer, pattern string, rest ...interface{}) (int, error) {
	return fmt.Fprintf(w, "%s", t.Func(pattern, rest...))
}
