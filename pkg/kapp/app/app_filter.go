// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import "time"

type AppFilter struct {
	CreatedAtBeforeTime *time.Time
	CreatedAtAfterTime  *time.Time

	Labels []string
}

func (f AppFilter) Apply(apps []App) ([]App, error) {
	var result []App

	for _, app := range apps {
		if f.Matches(app) {
			result = append(result, app)
		}
	}
	return result, nil
}

func (f AppFilter) Matches(app App) bool {
	if f.CreatedAtBeforeTime != nil {
		if app.CreationTimestamp().After(*f.CreatedAtBeforeTime) {
			return false
		}
	}

	if f.CreatedAtAfterTime != nil {
		if app.CreationTimestamp().Before(*f.CreatedAtAfterTime) {
			return false
		}
	}

	return true
}
