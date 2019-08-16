package core

import (
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
)

type MessagesUI struct {
	ui ui.UI
}

func NewMessagesUI(ui ui.UI) *MessagesUI {
	return &MessagesUI{ui: ui}
}

func (ui *MessagesUI) NotifySection(msg string, args ...interface{}) {
	ui.notify("---- "+msg+" ----", args...)
}

func (ui *MessagesUI) Notify(msgs []string) {
	for _, msg := range msgs {
		ui.notify("%s", msg)
	}
}

func (ui *MessagesUI) notify(msg string, args ...interface{}) {
	ui.ui.BeginLinef(time.Now().Format("3:04:05PM")+": "+msg+"\n", args...)
}
