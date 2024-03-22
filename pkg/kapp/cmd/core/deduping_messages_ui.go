// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	defaultForgetAfterDuration time.Duration = 1 * time.Minute
)

type DedupingMessagesUI struct {
	ui           MessagesUI
	lastSeen     map[string]time.Time
	lastSeenLock sync.Mutex
	forgetAfter  time.Duration
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

	if ui.shouldSeeAndMark(id) {
		ui.ui.NotifySection(msg)
	}
}

func (ui *DedupingMessagesUI) Notify(msgs []string) {
	id := strings.Join(msgs, "\n")

	if ui.shouldSeeAndMark(id) {
		ui.ui.Notify(msgs)
	}
}

func (ui *DedupingMessagesUI) shouldSeeAndMark(id string) bool {
	ui.lastSeenLock.Lock()
	defer ui.lastSeenLock.Unlock()

	when, found := ui.lastSeen[id]
	if !found || time.Now().Sub(when) > ui.forgetAfter {
		ui.lastSeen[id] = time.Now()
		return true
	}

	return false
}
