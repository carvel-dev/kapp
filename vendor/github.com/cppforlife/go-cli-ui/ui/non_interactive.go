package ui

import (
	. "github.com/cppforlife/go-cli-ui/ui/table"
)

type NonInteractiveUI struct {
	parent UI
}

func NewNonInteractiveUI(parent UI) *NonInteractiveUI {
	return &NonInteractiveUI{parent: parent}
}

func (ui *NonInteractiveUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *NonInteractiveUI) PrintLinef(pattern string, args ...interface{}) {
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *NonInteractiveUI) BeginLinef(pattern string, args ...interface{}) {
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *NonInteractiveUI) EndLinef(pattern string, args ...interface{}) {
	ui.parent.EndLinef(pattern, args...)
}

func (ui *NonInteractiveUI) PrintBlock(block []byte) {
	ui.parent.PrintBlock(block)
}

func (ui *NonInteractiveUI) PrintErrorBlock(block string) {
	ui.parent.PrintErrorBlock(block)
}

func (ui *NonInteractiveUI) PrintTable(table Table) {
	ui.parent.PrintTable(table)
}

func (ui *NonInteractiveUI) AskForText(label string) (string, error) {
	panic("Cannot ask for input in non-interactive UI")
}

func (ui *NonInteractiveUI) AskForChoice(label string, options []string) (int, error) {
	panic("Cannot ask for a choice in non-interactive UI")
}

func (ui *NonInteractiveUI) AskForPassword(label string) (string, error) {
	panic("Cannot ask for password in non-interactive UI")
}

func (ui *NonInteractiveUI) AskForConfirmation() error {
	// Always respond successfully
	return nil
}

func (ui *NonInteractiveUI) IsInteractive() bool {
	return false
}

func (ui *NonInteractiveUI) Flush() {
	ui.parent.Flush()
}
