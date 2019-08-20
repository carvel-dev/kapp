package util

type Throttle struct {
	ch chan struct{}
}

func NewThrottle(max int) Throttle {
	ch := make(chan struct{}, max)
	for i := 0; i < max; i++ {
		ch <- struct{}{}
	}
	return Throttle{ch}
}

func (t Throttle) Take() { <-t.ch }
func (t Throttle) Done() { t.ch <- struct{}{} }
