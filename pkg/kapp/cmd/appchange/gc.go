// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
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
	app, _, err := cmdapp.Factory(o.depsFactory, o.AppFlags, cmdapp.ResourceTypesFlags{}, o.logger, nil)
	if err != nil {
		return err
	}

	reviewFunc := func(changesToDelete []ctlapp.Change) error {
		AppChangesTable{"App changes to delete", changesToDelete}.Print(o.ui)

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
