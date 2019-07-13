package clusterapply

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
)

type ChangeSetViewOpts struct {
	Summary bool
	Changes bool
	ctldiff.TextDiffViewOpts
}

type ChangeSetView struct {
	changeViews []ChangeView
	opts        ChangeSetViewOpts

	changesView *ChangesView
}

func NewChangeSetView(changeViews []ChangeView, opts ChangeSetViewOpts) *ChangeSetView {
	return &ChangeSetView{changeViews, opts, nil}
}

func (v *ChangeSetView) Print(ui ui.UI) {
	if v.opts.Changes {
		for _, view := range v.changeViews {
			textDiffView := ctldiff.NewTextDiffView(view.TextDiff(), v.opts.TextDiffViewOpts)
			ui.BeginLinef("--- %s %s\n", view.ApplyOp(), view.Resource().Description())
			ui.PrintBlock([]byte(textDiffView.String()))
		}
	}

	v.changesView = &ChangesView{ChangeViews: v.changeViews, Sort: true}

	if v.opts.Summary {
		v.changesView.Print(ui)
	}
}

func (v *ChangeSetView) Summary() string {
	return v.changesView.Summary() // assumes Print was used before
}
