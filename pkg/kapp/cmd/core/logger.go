package core

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

type Logger struct {
	ui    ui.UI
	debug bool
}

func NewLogger(ui ui.UI) Logger {
	return NewLoggerWithDebug(ui, false)
}

func NewLoggerWithDebug(ui ui.UI, debug bool) Logger {
	return Logger{ui, debug}
}

func (l Logger) Error(tag string, msg string, args ...interface{}) {
	l.ui.BeginLinef(l.msg(loggerLevelError, tag, msg), args...)
}

func (l Logger) Info(tag string, msg string, args ...interface{}) {
	l.ui.BeginLinef(l.msg(loggerLevelInfo, tag, msg), args...)
}

func (l Logger) Debug(tag string, msg string, args ...interface{}) {
	if l.debug {
		l.ui.BeginLinef(l.msg(loggerLevelDebug, tag, msg), args...)
	}
}

func (l Logger) msg(level, tag, msg string) string {
	ts := time.Now().Format("03:04:05PM")
	return fmt.Sprintf("%s: %s: %s: %s\n", ts, level, tag, msg)
}
