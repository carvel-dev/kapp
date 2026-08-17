package ui

import (
	"fmt"

	. "github.com/cppforlife/go-cli-ui/ui/table"
)

type IndentingUI struct {
	parent UI
}

func NewIndentingUI(parent UI) *IndentingUI {
	return &IndentingUI{parent: parent}
}

func (ui *IndentingUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef("  %s", fmt.Sprintf(pattern, args...))
}

func (ui *IndentingUI) PrintLinef(pattern string, args ...interface{}) {
	ui.parent.PrintLinef("  %s", fmt.Sprintf(pattern, args...))
}

func (ui *IndentingUI) BeginLinef(pattern string, args ...interface{}) {
	ui.parent.BeginLinef("  %s", fmt.Sprintf(pattern, args...))
}

func (ui *IndentingUI) EndLinef(pattern string, args ...interface{}) {
	ui.parent.EndLinef(pattern, args...)
}

func (ui *IndentingUI) PrintBlock(block []byte) {
	ui.parent.PrintBlock(block)
}

func (ui *IndentingUI) PrintErrorBlock(block string) {
	ui.parent.PrintErrorBlock(block)
}

func (ui *IndentingUI) PrintTable(table Table) {
	ui.parent.PrintTable(table)
}

func (ui *IndentingUI) AskForText(opts TextOpts) (string, error) {
	return ui.parent.AskForText(opts)
}

func (ui *IndentingUI) AskForChoice(opts ChoiceOpts) (int, error) {
	return ui.parent.AskForChoice(opts)
}

func (ui *IndentingUI) AskForPassword(label string) (string, error) {
	return ui.parent.AskForPassword(label)
}

func (ui *IndentingUI) AskForConfirmation() error {
	return ui.parent.AskForConfirmation()
}

func (ui *IndentingUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *IndentingUI) Flush() {
	ui.parent.Flush()
}
