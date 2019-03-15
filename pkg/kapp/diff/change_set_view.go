package diff

import (
	"github.com/cppforlife/go-cli-ui/ui"
)

type ChangeSetViewOpts struct {
	Summary     bool
	SummaryFull bool
	Changes     bool
	ChangesFull bool
	TextDiffViewOpts
}

type ChangeSetView struct {
	changes []Change
	opts    ChangeSetViewOpts

	changesView *ChangesView
}

func NewChangeSetView(changes []Change, opts ChangeSetViewOpts) *ChangeSetView {
	return &ChangeSetView{changes, opts, nil}
}

func (v *ChangeSetView) Print(ui ui.UI) {
	if v.opts.Changes || v.opts.ChangesFull {
		for _, change := range v.changes {
			if v.opts.ChangesFull || v.isInterestingChange(change) {
				textDiffView := NewTextDiffView(change.TextDiff(), v.opts.TextDiffViewOpts)
				ui.BeginLinef("--- %s %s\n", change.Op(), change.NewOrExistingResource().Description())
				ui.PrintBlock([]byte(textDiffView.String()))
			}
		}
	}

	// Delegate filtering to the table view to be able to show full statistics
	includeChangeFunc := func(ch Change) bool {
		return v.opts.SummaryFull || v.isInterestingChange(ch)
	}

	v.changesView = &ChangesView{Changes: v.changes, Sort: true, IncludeFunc: includeChangeFunc}

	if v.opts.Summary || v.opts.SummaryFull {
		v.changesView.Print(ui)
	}
}

func (v *ChangeSetView) Summary() string {
	return v.changesView.Summary() // assumes Print was used before
}

func (ChangeSetView) isInterestingChange(change Change) bool {
	return change.Op() != ChangeOpKeep && !change.IsIgnored()
}
