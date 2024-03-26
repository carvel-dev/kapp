// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package logger

type NoopLogger struct{}

var _ Logger = NoopLogger{}

func NewNoopLogger() NoopLogger { return NoopLogger{} }
func NewTODOLogger() NoopLogger { return NewNoopLogger() }

func (l NoopLogger) Error(_ string, _ ...interface{}) {}
func (l NoopLogger) Info(_ string, _ ...interface{})  {}
func (l NoopLogger) Debug(_ string, _ ...interface{}) {}
func (l NoopLogger) DebugFunc(_ string) FuncLogger    { return NoopFuncLogger{} }
func (l NoopLogger) NewPrefixed(_ string) Logger      { return l }

type NoopFuncLogger struct{}

var _ FuncLogger = NoopFuncLogger{}

func (l NoopFuncLogger) Finish() {}
