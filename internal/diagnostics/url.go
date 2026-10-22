// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

// Package diagnostics formats endpoints without authentication values.
package diagnostics

import (
	"errors"
	"net/url"
)

// URL returns endpoint context without authentication or query/fragment values.
// The original URL must still be used for the actual request.
func URL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Opaque != "" {
		return "[invalid URL]"
	}
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	return u.String()
}

// Error removes URL wrappers which would otherwise reintroduce the raw URL.
func Error(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return Error(urlErr.Err)
	}
	return err
}
