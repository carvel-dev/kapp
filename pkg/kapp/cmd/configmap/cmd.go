package configmap

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config-map",
		Aliases: []string{"cm", "cfg", "config-maps", "configmap", "configmaps"},
		Short:   "Config map",
	}
	return cmd
}
