// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
)

const (
	loggerLevelError = "error"
	loggerLevelInfo  = "info"
	loggerLevelDebug = "debug"
)

type UILogger struct {
	prefix string
	ui     ui.UI
	debug  bool
}

var _ Logger = &UILogger{}

func NewUILogger(ui ui.UI) *UILogger { return &UILogger{"", ui, false} }

func (l *UILogger) SetDebug(debug bool) { l.debug = debug }

func (l *UILogger) Error(msg string, args ...interface{}) {
	l.ui.BeginLinef(l.msg(loggerLevelError, msg), args...)
}

func (l *UILogger) Info(msg string, args ...interface{}) {
	l.ui.BeginLinef(l.msg(loggerLevelInfo, msg), args...)
}

func (l *UILogger) Debug(msg string, args ...interface{}) {
	if l.debug {
		l.ui.BeginLinef(l.msg(loggerLevelDebug, msg), args...)
	}
}

func (l *UILogger) DebugFunc(name string) FuncLogger {
	funcLogger := &UIFuncLogger{name, time.Now(), l.NewPrefixed(name)}
	funcLogger.Start()
	return funcLogger
}

func (l *UILogger) NewPrefixed(name string) Logger {
	if len(l.prefix) > 0 {
		name = l.prefix + name
	}
	name += ": "
	return &UILogger{name, l.ui, l.debug}
}

func (l *UILogger) msg(level, msg string) string {
	ts := time.Now().Format("03:04:05PM")
	return fmt.Sprintf("%s: %s: %s%s\n", ts, level, l.prefix, msg)
}

type UIFuncLogger struct {
	name      string
	startTime time.Time
	logger    Logger
}

var _ FuncLogger = &UIFuncLogger{}

func (l *UIFuncLogger) Start()  { l.logger.Debug("start") }
func (l *UIFuncLogger) Finish() { l.logger.Debug("end (%s)", time.Now().Sub(l.startTime)) }
