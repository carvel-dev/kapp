// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/spf13/cobra"
)

const (
	appGroupAnnKey = "kapp.k14s.io/app-group"
)

type DeleteOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppGroupFlags Flags
	DeployFlags   DeployFlags
	AppFlags      DeleteAppFlags
}

type DeleteAppFlags struct {
	DiffFlags  cmdtools.DiffFlags
	ApplyFlags cmdapp.ApplyFlags
}

func NewDeleteOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *DeleteOptions {
	return &DeleteOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewDeleteCmd(o *DeleteOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete app group",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppGroupFlags.Set(cmd, flagsFactory)
	o.AppFlags.DiffFlags.SetWithPrefix("diff", cmd)
	o.AppFlags.ApplyFlags.SetWithDefaults("", cmdapp.ApplyFlagsDeleteDefaults, cmd)
	return cmd
}

func (o *DeleteOptions) Run() error {
	if len(o.AppGroupFlags.Name) == 0 {
		return fmt.Errorf("Expected group name to be non-empty")
	}

	supportObjs, err := cmdapp.FactoryClients(o.depsFactory, o.AppGroupFlags.NamespaceFlags, cmdapp.ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	appsInGroup, err := supportObjs.Apps.List(map[string]string{appGroupAnnKey: o.AppGroupFlags.Name})
	if err != nil {
		return err
	}

	for _, app := range appsInGroup {
		err := o.deleteApp(app.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *DeleteOptions) deleteApp(name string) error {
	o.ui.PrintLinef("--- deleting app '%s' (namespace: %s)",
		name, o.AppGroupFlags.NamespaceFlags.Name)

	deleteOpts := cmdapp.NewDeleteOptions(o.ui, o.depsFactory, o.logger)
	deleteOpts.AppFlags = cmdapp.Flags{
		Name:           name,
		NamespaceFlags: o.AppGroupFlags.NamespaceFlags,
	}
	deleteOpts.DiffFlags = o.AppFlags.DiffFlags
	deleteOpts.ApplyFlags = o.AppFlags.ApplyFlags

	return deleteOpts.Run()
}
