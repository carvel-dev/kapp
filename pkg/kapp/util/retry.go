// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	// "fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// Retry is different from wait.Poll because
// it does not stop retrying when error is encountered
func Retry(interval, timeout time.Duration, condFunc wait.ConditionFunc) error {
	var lastErr error
	var times int

	wait.PollImmediate(interval, timeout, func() (bool, error) {
		done, err := condFunc()
		lastErr = err
		times++
		return done, nil
	})

	if lastErr != nil {
		// TODO should not wrap error as it may lose necessary type info
		// eg resources.Update needs to return status info
		// return fmt.Errorf("Retried %d times: %s", times, lastErr)
		return lastErr
	}

	return nil
}

// Retry is different from wait.Poll because
// it does not stop retrying when error is encountered
func Retry2(interval, timeout time.Duration, shouldRetryFunc func(error) bool, performFunc func() error) error {
	var lastErr error

	wait.PollImmediate(interval, timeout, func() (bool, error) {
		err := performFunc()
		lastErr = err
		done := err == nil || shouldRetryFunc(err) == false
		return done, nil
	})

	if lastErr != nil {
		// TODO should not wrap error as it may lose necessary type info
		// eg resources.Update needs to return status info
		// return fmt.Errorf("Retried %d times: %s", times, lastErr)
		return lastErr
	}

	return nil
}
