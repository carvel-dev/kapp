package logger

type Logger interface {
	DebugFunc(name string) FuncLogger
	NewPrefixed(name string) Logger

	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type FuncLogger interface {
	Finish()
}
