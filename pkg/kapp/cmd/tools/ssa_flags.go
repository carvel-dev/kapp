package tools

import "github.com/spf13/cobra"

type SSAFlags struct {
	Enabled          bool
	ForceConflict    bool
	FieldManagerName string
}

func (s *SSAFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.Enabled, "ssa", false, "Use server side apply")
	cmd.Flags().StringVar(&s.FieldManagerName, "ssa-field-manager", "kapp-server-side-apply", "Name of the manager used to track field ownership")
	cmd.Flags().BoolVar(&s.ForceConflict, "ssa-force-conflicts", false, "If true, server-side apply will force the changes against conflicts.")
}
