// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"time"

	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	"github.com/spf13/cobra"
)

var (
	ApplyFlagsDeployDefaults = ApplyFlags{
		ClusterChangeOpts: ctlcap.ClusterChangeOpts{
			ApplyIgnored: false,
			Wait:         true,
			WaitIgnored:  false,
		},
	}
	ApplyFlagsDeleteDefaults = ApplyFlags{
		ClusterChangeOpts: ctlcap.ClusterChangeOpts{
			ApplyIgnored: false,
			Wait:         true,
			WaitIgnored:  true,
		},
	}
)

type ApplyFlags struct {
	ctlcap.ClusterChangeSetOpts
	ctlcap.ClusterChangeOpts

	ExitStatus bool
}

func (s *ApplyFlags) SetWithDefaults(prefix string, defaults ApplyFlags, cmd *cobra.Command) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.ApplyIgnored, prefix+"apply-ignored", defaults.ApplyIgnored, "Set to apply ignored changes")
	cmd.Flags().DurationVar(&s.ApplyingChangesOpts.Timeout, prefix+"apply-timeout",
		mustParseDuration("15m"), "Maximum amount of time to wait in apply phase")
	cmd.Flags().DurationVar(&s.ApplyingChangesOpts.CheckInterval, prefix+"apply-check-interval",
		mustParseDuration("1s"), "Amount of time to sleep between applies")
	cmd.Flags().IntVar(&s.ApplyingChangesOpts.Concurrency, prefix+"apply-concurrency", 5, "Maximum number of concurrent apply operations")

	cmd.Flags().StringVar(&s.AddOrUpdateChangeOpts.DefaultUpdateStrategy, prefix+"apply-default-update-strategy",
		defaults.AddOrUpdateChangeOpts.DefaultUpdateStrategy, "Change default update strategy")

	cmd.Flags().BoolVar(&s.Wait, prefix+"wait", defaults.Wait, "Set to wait for changes to be applied")
	cmd.Flags().BoolVar(&s.WaitIgnored, prefix+"wait-ignored", defaults.WaitIgnored, "Set to wait for ignored changes to be applied")

	cmd.Flags().DurationVar(&s.WaitingChangesOpts.Timeout, prefix+"wait-timeout",
		mustParseDuration("15m"), "Maximum amount of time to wait in wait phase")
	cmd.Flags().DurationVar(&s.WaitingChangesOpts.ResourceTimeout, prefix+"wait-resource-timeout",
		mustParseDuration("0"), "Maximum amount of time to wait for a resource in wait phase. 0s is the default value which indicates no timeout")
	cmd.Flags().DurationVar(&s.WaitingChangesOpts.CheckInterval, prefix+"wait-check-interval",
		mustParseDuration("1s"), "Amount of time to sleep between checks while waiting")
	cmd.Flags().IntVar(&s.WaitingChangesOpts.Concurrency, prefix+"wait-concurrency",
		5, "Maximum number of concurrent wait operations")

	cmd.Flags().BoolVar(&s.ExitStatus, prefix+"apply-exit-status", false, "Return specific exit status based on number of changes")
}

func mustParseDuration(str string) time.Duration {
	dur, err := time.ParseDuration(str)
	if err != nil {
		panic(fmt.Sprintf("Expected to successfully parse duration '%s': %s", str, err))
	}
	return dur
}
