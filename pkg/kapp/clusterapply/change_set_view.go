// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"github.com/cppforlife/color"
	"github.com/cppforlife/go-cli-ui/ui"
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	"github.com/k14s/kapp/pkg/kapp/diff"
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

func (v ChangeSetView) PrintCompleteYamlToBeApplied(ui ui.UI, conf ctlconf.Conf) error {
	for _, view := range v.changeViews {
		op := view.ApplyOp()
		opAndResDesc := ""
		switch op {
		case ClusterChangeApplyOpNoop:
			// will do nothing
		case ClusterChangeApplyOpDelete:
			strategy, _ := view.ApplyStrategyOp()
			if strategy == deleteStrategyPlainAnnValue {
				opAndResDesc = color.RedString("# %s: %s", view.ApplyOp(), view.Resource().Description())
			} else {
				opAndResDesc = color.RedString("# %s %s: %s", strategy, view.ApplyOp(), view.Resource().Description())
			}
			ui.PrintBlock([]byte(opAndResDesc + "\n"))
		case ClusterChangeApplyOpExists:
			opAndResDesc = color.GreenString("# %s: %s", view.ApplyOp(), view.Resource().Description())
			ui.PrintBlock([]byte(opAndResDesc + color.GreenString(" => kapp will wait for this resource to be created\n")))
		default:
			opAndResDesc = color.GreenString("# %s: %s", view.ApplyOp(), view.Resource().Description())
			resMgd := ctlres.NewResourceWithManagedFields(view.Resource(), false)
			res, err := resMgd.Resource()
			if err != nil {
				return err
			}

			res, err = diff.NewMaskedResource(res, conf.DiffMaskRules()).Resource()
			if err != nil {
				return err
			}

			resYAML, err := res.AsYAMLBytes()
			if err != nil {
				return err
			}

			ui.PrintBlock([]byte(opAndResDesc + "\n" + "---\n" + string(resYAML)))
		}
	}
	return nil
}
