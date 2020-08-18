// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

type DoneApplyState struct {
	Done       bool
	Successful bool
	Message    string
}

func (s DoneApplyState) TerminallyFailed() bool {
	return s.Done && !s.Successful
}
