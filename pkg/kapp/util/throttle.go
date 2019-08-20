package util

import (
	"fmt"
)

type Throttle struct {
	ch chan struct{}
}

func NewThrottle(max int) Throttle {
	if max < 1 {
		panic(fmt.Sprintf("Expected maximum throttle to be >= 1, but was %d", max))
	}
	ch := make(chan struct{}, max)
	for i := 0; i < max; i++ {
		ch <- struct{}{}
	}
	return Throttle{ch}
}

func (t Throttle) Take() { <-t.ch }
func (t Throttle) Done() { t.ch <- struct{}{} }
