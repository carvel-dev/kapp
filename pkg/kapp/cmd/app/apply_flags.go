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
}

func (s *ApplyFlags) SetWithDefaults(prefix string, defaults ApplyFlags, cmd *cobra.Command) {
	if len(prefix) > 0 {
		prefix += "-"
	}

	cmd.Flags().BoolVar(&s.ApplyIgnored, prefix+"apply-ignored", defaults.ApplyIgnored, "Set to apply ignored changes")
	cmd.Flags().BoolVar(&s.Wait, prefix+"apply-wait", defaults.Wait, "Set to wait for changes to be applied")
	cmd.Flags().BoolVar(&s.WaitIgnored, prefix+"apply-wait-ignored", defaults.WaitIgnored, "Set to wait for ignored changes to be applied")

	cmd.Flags().DurationVar(&s.WaitTimeout, prefix+"apply-wait-timeout",
		mustParseDuration("15m"), "Maximum amount of time to wait")
	cmd.Flags().DurationVar(&s.WaitCheckInterval, prefix+"apply-wait-check-interval",
		mustParseDuration("1s"), "Amount of time to sleep between checks while waiting")

	cmd.Flags().StringVar(&s.AddOrUpdateChangeOpts.DefaultUpdateStrategy, prefix+"apply-default-update-strategy",
		defaults.AddOrUpdateChangeOpts.DefaultUpdateStrategy, "Change default update strategy")
}

func mustParseDuration(str string) time.Duration {
	dur, err := time.ParseDuration(str)
	if err != nil {
		panic(fmt.Sprintf("Expected to successfully parse duration '%s': %s", str, err))
	}
	return dur
}
