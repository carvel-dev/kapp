package core

import (
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
)

type MessagesUI struct {
	ui ui.UI
}

func NewMessagesUI(ui ui.UI) MessagesUI {
	return MessagesUI{ui}
}

func (ui MessagesUI) Notify(msg string, args ...interface{}) {
	ui.NotifyBegin(msg, args...)
	ui.NotifyEnd("")
}

func (ui MessagesUI) NotifyBegin(msg string, args ...interface{}) {
	ui.ui.BeginLinef(time.Now().Format("3:04:05PM")+": "+msg, args...)
}

func (ui MessagesUI) NotifyEnd(msg string, args ...interface{}) {
	ui.ui.BeginLinef(msg+"\n", args...)
}
