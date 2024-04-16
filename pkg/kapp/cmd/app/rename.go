// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type RenameOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags     Flags
	NewName      string
	NewNamespace string
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
	cmd.Flags().StringVar(&o.NewNamespace, "new-namespace", "", "Set new namespace (format: new-namespace)")
	return cmd
}

func (o *RenameOptions) Run() error {
	app, _, err := Factory(o.depsFactory, o.AppFlags, ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	exists, notExistsMsg, err := app.Exists()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s", notExistsMsg)
	}

	newName := o.NewName
	newNamespace := o.NewNamespace
	if newName == "" && newNamespace == "" {
		return fmt.Errorf("Expected either --new-name or/and --new-namespace to be supplied")
	}

	if newName == "" {
		newName = app.Name()
	}

	if newNamespace == "" {
		newNamespace = o.AppFlags.NamespaceFlags.Name
	}

	o.ui.PrintLinef("Renaming '%s' (namespace: %s) to '%s' (namespace: %s) (app changes will not be renamed)",
		app.Name(), o.AppFlags.NamespaceFlags.Name, newName, newNamespace)

	err = o.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return app.Rename(newName, newNamespace)
}
