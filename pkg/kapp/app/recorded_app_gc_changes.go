// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

const (
	AppChangesMaxToKeepDefault = 200
)

func (a *RecordedApp) GCChanges(max int, reviewFunc func(changesToDelete []Change) error) (int, int, error) {
	if reviewFunc == nil {
		reviewFunc = func(_ []Change) error { return nil }
	}

	changes, err := a.Changes()
	if err != nil {
		return 0, 0, err
	}

	if len(changes) < max {
		return len(changes), 0, reviewFunc(nil)
	}

	// First change is oldest
	changes = changes[0 : len(changes)-max]

	err = reviewFunc(changes)
	if err != nil {
		return 0, 0, err
	}

	for _, change := range changes {
		err := change.Delete()
		if err != nil {
			return 0, 0, err
		}
	}

	return max, len(changes), nil
}
