// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	"fmt"

	cmdapp "carvel.dev/kapp/pkg/kapp/cmd/app"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	cmdtools "carvel.dev/kapp/pkg/kapp/cmd/tools"
	"carvel.dev/kapp/pkg/kapp/logger"
	"github.com/cppforlife/go-cli-ui/ui"
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
		Use:         "delete",
		Short:       "Delete app group",
		RunE:        func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{cmdapp.TTYByDefaultKey: ""},
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

	supportObjs, err := cmdapp.FactoryClients(o.depsFactory, o.AppGroupFlags.NamespaceFlags, o.AppGroupFlags.AppNamespace, cmdapp.ResourceTypesFlags{}, o.logger)
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
		name, o.appNamespace())

	deleteOpts := cmdapp.NewDeleteOptions(o.ui, o.depsFactory, o.logger)
	deleteOpts.AppFlags = cmdapp.Flags{
		Name:           name,
		NamespaceFlags: o.AppGroupFlags.NamespaceFlags,
		AppNamespace:   o.AppGroupFlags.AppNamespace,
	}
	deleteOpts.DiffFlags = o.AppFlags.DiffFlags
	deleteOpts.ApplyFlags = o.AppFlags.ApplyFlags

	return deleteOpts.Run()
}

func (o *DeleteOptions) appNamespace() string {
	if o.AppGroupFlags.AppNamespace != "" {
		return o.AppGroupFlags.AppNamespace
	}
	return o.AppGroupFlags.NamespaceFlags.Name
}
