// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

type DoneApplyState struct {
	Done       bool   `json:"done"`
	Successful bool   `json:"successful"`
	Message    string `json:"message"`

	UnblockChanges bool `json:"unblockChanges"`
}

func (s DoneApplyState) TerminallyFailed() bool {
	return s.Done && !s.Successful
}
