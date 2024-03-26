// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"github.com/cppforlife/cobrautil"
)

const (
	cmdGroupKey = "kapp-group"
)

var (
	AppHelpGroup = cobrautil.HelpSection{
		Key:   cmdGroupKey,
		Value: "app",
		Title: "App Commands:",
	}
	AppSupportHelpGroup = cobrautil.HelpSection{
		Key:   cmdGroupKey,
		Value: "app-support",
		Title: "App Support Commands:",
	}
	MiscHelpGroup = cobrautil.HelpSection{
		Key:   cmdGroupKey,
		Value: "misc",
		Title: "Misc Commands:",
	}
	RestOfCommandsHelpGroup = cobrautil.HelpSection{
		Key:   cmdGroupKey,
		Value: "", // default
		Title: "Available/Other Commands:",
	}
)
