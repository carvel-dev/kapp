// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

type NoopLogger struct{}

var _ Logger = NoopLogger{}

func NewNoopLogger() NoopLogger { return NoopLogger{} }
func NewTODOLogger() NoopLogger { return NewNoopLogger() }

func (l NoopLogger) Error(msg string, args ...interface{}) {}
func (l NoopLogger) Info(msg string, args ...interface{})  {}
func (l NoopLogger) Debug(msg string, args ...interface{}) {}
func (l NoopLogger) DebugFunc(name string) FuncLogger      { return NoopFuncLogger{} }
func (l NoopLogger) NewPrefixed(name string) Logger        { return l }

type NoopFuncLogger struct{}

var _ FuncLogger = NoopFuncLogger{}

func (l NoopFuncLogger) Finish() {}
