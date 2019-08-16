package clusterapply

type UI interface {
	NotifySection(msg string, args ...interface{})
	Notify(msgs []string)
}
