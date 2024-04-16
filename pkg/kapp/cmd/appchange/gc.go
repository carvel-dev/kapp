// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	ctlapp "carvel.dev/kapp/pkg/kapp/app"
	cmdapp "carvel.dev/kapp/pkg/kapp/cmd/app"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type GCOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags cmdapp.Flags
	Max      int
}

func NewGCOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *GCOptions {
	return &GCOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewGCCmd(o *GCOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "garbage-collect",
		Aliases: []string{"gc"},
		Short:   "Garbage collect app changes",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	cmd.Flags().IntVar(&o.Max, "max", ctlapp.AppChangesMaxToKeepDefault, "Maximum number of app changes to keep")
	return cmd
}

func (o *GCOptions) Run() error {
	app, _, err := cmdapp.Factory(o.depsFactory, o.AppFlags, cmdapp.ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	reviewFunc := func(changesToDelete []ctlapp.Change) error {
		AppChangesTable{"App changes to delete", changesToDelete, TimeFlags{}}.Print(o.ui)

		err = o.ui.AskForConfirmation()
		if err != nil {
			return err
		}

		return nil
	}

	numKept, numDeleted, err := app.GCChanges(o.Max, reviewFunc)
	if err != nil {
		return err
	}

	o.ui.PrintLinef("Kept %d, deleted %d app changes", numKept, numDeleted)
	return nil
}
