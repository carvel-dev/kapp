package ui

type NoopLogger struct{}

func NewNoopLogger() NoopLogger {
	return NoopLogger{}
}

var _ ExternalLogger = NoopLogger{}

func (l NoopLogger) Error(tag, msg string, args ...interface{}) {}
func (l NoopLogger) Debug(tag, msg string, args ...interface{}) {}
