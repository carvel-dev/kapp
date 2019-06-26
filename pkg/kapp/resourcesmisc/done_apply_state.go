package resourcesmisc

type DoneApplyState struct {
	Done       bool
	Successful bool
	Message    string
}

func (s DoneApplyState) TerminallyFailed() bool {
	return s.Done && !s.Successful
}
