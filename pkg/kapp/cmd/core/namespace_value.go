// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

func NewValueNamespace(ns string) uitable.ValueString {
	if len(ns) > 0 {
		return uitable.NewValueString(ns)
	}
	return uitable.NewValueString("(cluster)")
}
