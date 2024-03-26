// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"sync"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
)

type MessagesUI interface {
	NotifySection(msg string, args ...interface{})
	Notify(msgs []string)
}

type PlainMessagesUI struct {
	ui     ui.UI
	uiLock sync.RWMutex
}

var _ MessagesUI = &PlainMessagesUI{}

func NewPlainMessagesUI(ui ui.UI) *PlainMessagesUI {
	return &PlainMessagesUI{ui: ui}
}

func (ui *PlainMessagesUI) NotifySection(msg string, args ...interface{}) {
	ui.notify("---- "+msg+" ----", args...)
}

func (ui *PlainMessagesUI) Notify(msgs []string) {
	for _, msg := range msgs {
		ui.notify("%s", msg)
	}
}

func (ui *PlainMessagesUI) notify(msg string, args ...interface{}) {
	ui.uiLock.Lock()
	defer ui.uiLock.Unlock()

	ui.ui.BeginLinef(time.Now().Format("3:04:05PM")+": "+msg+"\n", args...)
}
