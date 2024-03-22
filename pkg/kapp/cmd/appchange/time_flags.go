// Copyright 2024 The Carvel Authors.
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
	cmd.Flags().StringVar(&t.Before, "before", "", "List app changes before given time (formats: 2022-01-10T12:15:25Z, 2022-01-10)")
	cmd.Flags().StringVar(&t.After, "after", "", "List app changes after given time (formats: 2022-01-10T12:15:25Z, 2022-01-10)")
}
