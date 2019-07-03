package core

import (
	"sync"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
)

type MessagesUI struct {
	ui                  ui.UI
	sawFirstSection     bool
	sawFirstSectionLock sync.Mutex
}

func NewMessagesUI(ui ui.UI) *MessagesUI {
	return &MessagesUI{ui: ui, sawFirstSection: false}
}

func (ui *MessagesUI) NotifySection(msg string, args ...interface{}) {
	ui.sawFirstSectionLock.Lock()
	defer ui.sawFirstSectionLock.Unlock()

	if ui.sawFirstSection {
		ui.ui.BeginLinef("\n")
	} else {
		ui.sawFirstSection = true
	}

	ui.notify("---- "+msg+" ----", args...)
}

func (ui *MessagesUI) Notify(msg string, args ...interface{}) {
	ui.notify(msg, args...)
}

func (ui *MessagesUI) notify(msg string, args ...interface{}) {
	ui.ui.BeginLinef(time.Now().Format("3:04:05PM")+": "+msg+"\n", args...)
}
