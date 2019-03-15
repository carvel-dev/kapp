package ui

import (
	. "github.com/cppforlife/go-cli-ui/ui/table"
)

type paddingUIMode int

const (
	paddingUIModeNone paddingUIMode = iota
	paddingUIModeRaw
	paddingUIModeAuto
	paddingUIModeAskText
)

type PaddingUI struct {
	parent   UI
	prevMode paddingUIMode
}

func NewPaddingUI(parent UI) *PaddingUI {
	return &PaddingUI{parent: parent}
}

func (ui *PaddingUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeAuto)
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *PaddingUI) PrintLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeAuto)
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *PaddingUI) BeginLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *PaddingUI) EndLinef(pattern string, args ...interface{}) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.EndLinef(pattern, args...)
}

func (ui *PaddingUI) PrintBlock(block []byte) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.PrintBlock(block)
}

func (ui *PaddingUI) PrintErrorBlock(block string) {
	ui.padBefore(paddingUIModeRaw)
	ui.parent.PrintErrorBlock(block)
}

func (ui *PaddingUI) PrintTable(table Table) {
	ui.padBefore(paddingUIModeAuto)
	ui.parent.PrintTable(table)
}

func (ui *PaddingUI) AskForText(label string) (string, error) {
	ui.padBefore(paddingUIModeAskText)
	return ui.parent.AskForText(label)
}

func (ui *PaddingUI) AskForChoice(label string, options []string) (int, error) {
	ui.padBefore(paddingUIModeAuto)
	return ui.parent.AskForChoice(label, options)
}

func (ui *PaddingUI) AskForPassword(label string) (string, error) {
	ui.padBefore(paddingUIModeAskText)
	return ui.parent.AskForPassword(label)
}

func (ui *PaddingUI) AskForConfirmation() error {
	ui.padBefore(paddingUIModeAuto)
	return ui.parent.AskForConfirmation()
}

func (ui *PaddingUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *PaddingUI) Flush() {
	ui.parent.Flush()
}

func (ui *PaddingUI) padBefore(currMode paddingUIMode) {
	switch {
	case ui.prevMode == paddingUIModeNone:
		// do nothing on the first time UI is called
	case ui.prevMode == paddingUIModeAskText && currMode == paddingUIModeAskText:
		// do nothing
	case ui.prevMode == paddingUIModeRaw && currMode == paddingUIModeRaw:
		// do nothing
	default:
		ui.parent.PrintLinef("")
	}
	ui.prevMode = currMode
}
