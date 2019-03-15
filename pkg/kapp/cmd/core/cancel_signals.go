package core

import (
	"os"
	"os/signal"
	"syscall"
)

type CancelSignals struct{}

func (CancelSignals) Watch(stopFunc func()) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		defer signal.Stop(signalCh)
		select {
		case <-signalCh:
			stopFunc()
		}
	}()
}
