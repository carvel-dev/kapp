package ui

import (
	. "github.com/cppforlife/go-cli-ui/ui/table"
)

type NonTTYUI struct {
	parent UI
}

func NewNonTTYUI(parent UI) *NonTTYUI {
	return &NonTTYUI{parent: parent}
}

func (ui *NonTTYUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *NonTTYUI) PrintLinef(pattern string, args ...interface{}) {}
func (ui *NonTTYUI) BeginLinef(pattern string, args ...interface{}) {}
func (ui *NonTTYUI) EndLinef(pattern string, args ...interface{})   {}

func (ui *NonTTYUI) PrintBlock(block []byte)      { ui.parent.PrintBlock(block) }
func (ui *NonTTYUI) PrintErrorBlock(block string) { ui.parent.PrintErrorBlock(block) }

func (ui *NonTTYUI) PrintTable(table Table) {
	// hide decorations
	table.Title = ""
	table.Notes = nil
	table.Content = ""
	table.DataOnly = true

	// necessary for grep
	table.FillFirstColumn = true

	// cut's default delim
	table.BorderStr = "\t"

	ui.parent.PrintTable(table)
}

func (ui *NonTTYUI) AskForText(opts TextOpts) (string, error) {
	return ui.parent.AskForText(opts)
}

func (ui *NonTTYUI) AskForChoice(opts ChoiceOpts) (int, error) {
	return ui.parent.AskForChoice(opts)
}

func (ui *NonTTYUI) AskForPassword(label string) (string, error) {
	return ui.parent.AskForPassword(label)
}

func (ui *NonTTYUI) AskForConfirmation() error {
	return ui.parent.AskForConfirmation()
}

func (ui *NonTTYUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *NonTTYUI) Flush() {
	ui.parent.Flush()
}
