package clusterapply

type UI interface {
	Notify(msg string, args ...interface{})
	NotifyBegin(msg string, args ...interface{})
	NotifyEnd(msg string, args ...interface{})
}
