package app

import (
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	"github.com/spf13/cobra"
)

var (
	ApplyFlagsDeployDefaults = ApplyFlags{
		ctlcap.ClusterChangeOpts{
			ApplyIgnored: false,
			Wait:         true,
			WaitIgnored:  false,
		},
	}
	ApplyFlagsDeleteDefaults = ApplyFlags{
		ctlcap.ClusterChangeOpts{
			ApplyIgnored: false,
			Wait:         true,
			WaitIgnored:  true,
		},
	}
)

type ApplyFlags struct {
	ctlcap.ClusterChangeOpts
}

func (s *ApplyFlags) SetWithDefaults(defaults ApplyFlags, cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.ApplyIgnored, "apply-ignored", defaults.ApplyIgnored, "Set to apply ignored changes")
	cmd.Flags().BoolVar(&s.Wait, "apply-wait", defaults.Wait, "Set to wait for changes to be applied")
	cmd.Flags().BoolVar(&s.WaitIgnored, "apply-wait-ignored", defaults.WaitIgnored, "Set to wait for ignored changes to be applied")

	cmd.Flags().StringVar(&s.AddOrUpdateChangeOpts.DefaultUpdateStrategy, "apply-default-update-strategy",
		defaults.AddOrUpdateChangeOpts.DefaultUpdateStrategy, "Change default update strategy")
}
