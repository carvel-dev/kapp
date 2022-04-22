// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

type DoneApplyState struct {
	Done       bool   `json:"done"`
	Successful bool   `json:"successful"`
	Message    string `json:"message"`

	UnblockBlockedChanges bool
}

func (s DoneApplyState) TerminallyFailed() bool {
	return s.Done && !s.Successful
}
