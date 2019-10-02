package app

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/spf13/cobra"
)

type RenameOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags AppFlags
	NewName  string
}

func NewRenameOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *RenameOptions {
	return &RenameOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewRenameCmd(o *RenameOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename app",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	o.AppFlags.Set(cmd, flagsFactory)
	cmd.Flags().StringVar(&o.NewName, "new-name", "", "Set new name (format: new-name)")
	return cmd
}

func (o *RenameOptions) Run() error {
	app, _, _, err := AppFactory(o.depsFactory, o.AppFlags, o.logger)
	if err != nil {
		return err
	}

	exists, err := app.Exists()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("App '%s' (namespace: %s) does not exist", app.Name(), o.AppFlags.NamespaceFlags.Name)
	}

	o.ui.PrintLinef("Renaming '%s' (namespace: %s) to '%s' (app changes will not be renamed)",
		app.Name(), o.AppFlags.NamespaceFlags.Name, o.NewName)

	err = o.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return app.Rename(o.NewName)
}
