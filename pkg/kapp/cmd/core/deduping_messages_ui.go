package core

import (
	"fmt"
	"strings"
	"time"
)

const (
	defaultForgetAfterDuration time.Duration = 1 * time.Minute
)

type DedupingMessagesUI struct {
	ui          MessagesUI
	lastSeen    map[string]time.Time
	forgetAfter time.Duration
}

var _ MessagesUI = &DedupingMessagesUI{}

func NewDedupingMessagesUI(ui MessagesUI) *DedupingMessagesUI {
	return &DedupingMessagesUI{
		ui:          ui,
		lastSeen:    map[string]time.Time{},
		forgetAfter: defaultForgetAfterDuration,
	}
}

func (ui *DedupingMessagesUI) NotifySection(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	id := msg

	if ui.shouldSee(id) {
		ui.markSeen(id)
		ui.ui.NotifySection(msg)
	}
}

func (ui *DedupingMessagesUI) Notify(msgs []string) {
	id := strings.Join(msgs, "\n")

	if ui.shouldSee(id) {
		ui.markSeen(id)
		ui.ui.Notify(msgs)
	}
}

func (ui *DedupingMessagesUI) shouldSee(id string) bool {
	when, found := ui.lastSeen[id]
	if !found {
		return true
	}
	return time.Now().Sub(when) > ui.forgetAfter
}

func (ui *DedupingMessagesUI) markSeen(id string) {
	ui.lastSeen[id] = time.Now()
}
