// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeSetViewOpts struct {
	Summary bool
	Changes bool
	ctldiff.TextDiffViewOpts
}

type ChangeSetView struct {
	changeViews []ChangeView
	maskRules   []ctlconf.DiffMaskRule
	opts        ChangeSetViewOpts

	changesView *ChangesView
}

func NewChangeSetView(changeViews []ChangeView,
	maskRules []ctlconf.DiffMaskRule, opts ChangeSetViewOpts) *ChangeSetView {

	return &ChangeSetView{changeViews, maskRules, opts, nil}
}

func (v *ChangeSetView) Print(ui ui.UI) {
	if v.opts.Changes {
		for _, view := range v.changeViews {
			textDiffView := ctldiff.NewTextDiffView(view.ConfigurableTextDiff(), v.maskRules, v.opts.TextDiffViewOpts)
			ui.BeginLinef("@@ %s %s @@\n", applyOpCodeUI[view.ApplyOp()], view.Resource().Description())
			ui.PrintBlock([]byte(textDiffView.String()))
		}
	}

	v.changesView = &ChangesView{ChangeViews: v.changeViews, Sort: true, countsView: NewChangesCountsView()}

	if v.opts.Summary {
		v.changesView.Print(ui)
	}
}

func (v *ChangeSetView) Summary() string {
	return v.changesView.Summary() // assumes Print was used before
}

func (v ChangeSetView) PrintCompleteYamlToBeApplied(ui ui.UI) {

	for _, view := range v.changeViews {
		st, _ := view.ApplyStrategyOp()
		if view.ApplyOp() == ClusterChangeApplyOpDelete && st == deleteStrategyPlainAnnValue {
			continue
		}
		resMgd := ctlres.NewResourceWithManagedFields(view.Resource(), false)
		res, err := resMgd.Resource()
		if err != nil {
			fmt.Errorf("Error: [%s]\n", err.Error())
			return
		}
		by, err := res.AsYAMLBytes()
		if err != nil {
			fmt.Errorf("Error: [%s]\n", err.Error())
			return
		}
		ui.PrintBlock([]byte("---\n"))
		ui.PrintBlock(by)
	}
}
