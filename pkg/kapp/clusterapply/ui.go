// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
)

type UI interface {
	NotifySection(msg string, args ...interface{})
	Notify(msgs []string)
}

type DoneApplyStateUI struct {
	State   string
	Message string
	Error   bool
}

func NewDoneApplyStateUI(state ctlresm.DoneApplyState, err error) DoneApplyStateUI {
	if err != nil {
		return DoneApplyStateUI{State: "error", Message: err.Error(), Error: true}
	}
	switch {
	case state.Done && state.Successful:
		return DoneApplyStateUI{State: "ok", Message: state.Message, Error: false}
	case state.Done && !state.Successful:
		return DoneApplyStateUI{State: "fail", Message: state.Message, Error: true}
	case !state.Done:
		return DoneApplyStateUI{State: "ongoing", Message: state.Message, Error: true}
	default:
		return DoneApplyStateUI{State: "unknown", Message: state.Message, Error: true}
	}
}
