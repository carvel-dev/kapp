// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	"fmt"

	"github.com/cppforlife/color"
	"github.com/cppforlife/go-cli-ui/ui"
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	"github.com/k14s/kapp/pkg/kapp/diff"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeSetViewOpts struct {
	Summary     bool
	Changes     bool
	ChangesYAML bool
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
	if v.opts.ChangesYAML {
		v.printChangesYAML(ui)
	}
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

func (v ChangeSetView) printChangesYAML(ui ui.UI) error {
	for _, view := range v.changeViews {
		resYAML := ""
		opAndResDesc := fmt.Sprintf("# %s: %s", applyOpCodeUI[view.ApplyOp()], view.Resource().Description())
		strategy, err := view.ApplyStrategyOp()
		if err != nil {
			return err
		}

		switch view.ApplyOp() {
		case ClusterChangeApplyOpNoop:
			continue
		case ClusterChangeApplyOpDelete:
			if strategy == deleteStrategyPlainAnnValue {
				opAndResDesc = color.RedString(opAndResDesc)
			} else {
				opAndResDesc = color.RedString("%s (strategy: %s)", opAndResDesc, strategy)
			}
		case ClusterChangeApplyOpExists:
			opAndResDesc = color.GreenString(opAndResDesc)
		default:
			if strategy != createStrategyPlainAnnValue && strategy != updateStrategyPlainAnnValue {
				opAndResDesc = color.GreenString("%s (strategy: %s)", opAndResDesc, strategy)
			} else {
				opAndResDesc = color.GreenString(opAndResDesc)
			}
			res, err := ctlres.NewResourceWithManagedFields(view.Resource(), false).Resource()
			if err != nil {
				return err
			}

			res, err = diff.NewMaskedResource(res, v.maskRules).Resource()
			if err != nil {
				return err
			}

			resBytes, err := res.AsYAMLBytes()
			if err != nil {
				return err
			}
			resYAML = string(resBytes)
		}
		ui.PrintBlock([]byte(fmt.Sprintf(`---
%s
%s
`, opAndResDesc, resYAML)))
	}
	return nil
}
