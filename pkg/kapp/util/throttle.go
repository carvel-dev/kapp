// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
)

type Throttle struct {
	ch chan struct{}
}

func NewThrottle(maxValue int) Throttle {
	if maxValue < 1 {
		panic(fmt.Sprintf("Expected maximum throttle to be >= 1, but was %d", maxValue))
	}
	ch := make(chan struct{}, maxValue)
	for i := 0; i < maxValue; i++ {
		ch <- struct{}{}
	}
	return Throttle{ch}
}

func (t Throttle) Take() { <-t.ch }
func (t Throttle) Done() { t.ch <- struct{}{} }
