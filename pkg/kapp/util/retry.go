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

// Copied from newer version of "k8s.io/client-go/util/retry" OnError

// RetryOnError allows the caller to retry fn in case the error returned by fn is retriable
// according to the provided function. backoff defines the maximum retries and the wait
// interval between two retries.
func RetryOnError(backoff wait.Backoff, retriable func(error) bool, fn func() error) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case retriable(err):
			lastErr = err
			return false, nil
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return err
}
