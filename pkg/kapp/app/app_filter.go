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
		lastChange, err := app.LastChange()
		if err != nil {
			return []App{}, err
		}

		if f.Matches(lastChange) {
			result = append(result, app)
		}
	}
	return result, nil
}

func (f AppFilter) Matches(change Change) bool {

	if f.CreatedAtBeforeTime != nil {
		if change.Meta().StartedAt.After(*f.CreatedAtBeforeTime) {
			return false
		}
	}

	if f.CreatedAtAfterTime != nil {
		if change.Meta().StartedAt.Before(*f.CreatedAtAfterTime) {
			return false
		}
	}
	return true
}
