// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	"time"

	"github.com/spf13/cobra"
)

type TimeFlags struct {
	Before     string
	After      string
	BeforeTime time.Time
	AfterTime  time.Time

	Duration string
}

func (t *TimeFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&t.Before, "before", "", "", "List app-changes happened before given time stamp (format: before=2022-01-10T12:15:25Z)")
	cmd.Flags().StringVarP(&t.After, "after", "", "", "List app-changes happened after given time stamp (format: after=2022-01-10T12:15:25Z)")
}
