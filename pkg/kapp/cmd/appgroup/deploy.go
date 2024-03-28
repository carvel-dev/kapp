// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appgroup

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	cmdapp "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/tools"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/logger"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/preflight"
)

type DeployOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppGroupFlags   Flags
	DeployFlags     DeployFlags
	AppFlags        DeployAppFlags
	PreflightChecks *preflight.Registry
}

type DeployAppFlags struct {
	DiffFlags           cmdtools.DiffFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	ApplyFlags          cmdapp.ApplyFlags
	DeleteApplyFlags    cmdapp.ApplyFlags
	DeployFlags         cmdapp.DeployFlags
	LabelFlags          cmdapp.LabelFlags
}

func NewDeployOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger, preflights *preflight.Registry) *DeployOptions {
	return &DeployOptions{ui: ui, depsFactory: depsFactory, logger: logger, PreflightChecks: preflights}
}

func NewDeployCmd(o *DeployOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "deploy",
		Aliases:     []string{"d", "dep"},
		Short:       "Deploy app group",
		RunE:        func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{cmdapp.TTYByDefaultKey: ""},
	}
	o.AppGroupFlags.Set(cmd, flagsFactory)
	o.DeployFlags.Set(cmd)
	o.AppFlags.DiffFlags.SetWithPrefix("diff", cmd)
	o.AppFlags.ResourceFilterFlags.Set(cmd)
	o.AppFlags.ApplyFlags.SetWithDefaults("", cmdapp.ApplyFlagsDeployDefaults, cmd)
	o.AppFlags.DeleteApplyFlags.SetWithDefaults("delete", cmdapp.ApplyFlagsDeleteDefaults, cmd)
	o.AppFlags.DeployFlags.Set(cmd)
	o.AppFlags.LabelFlags.Set(cmd)
	o.PreflightChecks.AddFlags(cmd.Flags())
	return cmd
}

func (o *DeployOptions) Run() error {
	if len(o.AppGroupFlags.Name) == 0 {
		return fmt.Errorf("Expected group name to be non-empty")
	}

	// TODO what if app is renamed? currently it
	// will have conflicting resources with new-named app
	updatedApps, err := o.appsToUpdate()
	if err != nil {
		return err
	}

	var exitCode float64
	// TODO is there some order between apps?
	for _, appGroupApp := range updatedApps {
		err := o.deployApp(appGroupApp)
		if err != nil {
			if deployErr, ok := err.(cmdapp.DeployDiffExitStatus); ok {
				exitCode = math.Max(exitCode, float64(deployErr.ExitStatus()))
			} else {
				return err
			}
		}
	}

	supportObjs, err := cmdapp.FactoryClients(o.depsFactory, o.AppGroupFlags.NamespaceFlags, o.AppGroupFlags.AppNamespace, cmdapp.ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	existingAppsInGroup, err := supportObjs.Apps.List(map[string]string{appGroupAnnKey: o.AppGroupFlags.Name})
	if err != nil {
		return err
	}

	// Delete apps that no longer are present in directories
	for _, app := range existingAppsInGroup {
		var found bool
		for _, v := range updatedApps {
			if app.Name() == v.Name {
				found = true
				break
			}
		}
		if !found {
			err := o.deleteApp(app.Name())
			if err != nil {
				return err
			}
		}
	}

	if o.AppFlags.DiffFlags.Run && o.AppFlags.DiffFlags.ExitStatus {
		var hasNoChanges = exitCode == 2
		return cmdapp.DeployDiffExitStatus{HasNoChanges: hasNoChanges}
	}

	return nil
}

type appGroupApp struct {
	Name string
	Path string
}

func (o *DeployOptions) appsToUpdate() ([]appGroupApp, error) {
	var applications []appGroupApp

	dir := o.DeployFlags.Directory

	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Reading directory '%s': %w", dir, err)
	}

	for _, fi := range fileInfos {
		if !fi.IsDir() {
			continue
		}
		app := appGroupApp{
			Name: fmt.Sprintf("%s-%s", o.AppGroupFlags.Name, fi.Name()),
			Path: filepath.Join(dir, fi.Name()),
		}
		applications = append(applications, app)
	}
	return applications, nil
}

func (o *DeployOptions) deployApp(app appGroupApp) error {
	o.ui.PrintLinef("--- deploying app '%s' (namespace: %s) from %s",
		app.Name, o.appNamespace(), app.Path)

	deployOpts := cmdapp.NewDeployOptions(o.ui, o.depsFactory, o.logger, o.PreflightChecks)
	deployOpts.AppFlags = cmdapp.Flags{
		Name:           app.Name,
		NamespaceFlags: o.AppGroupFlags.NamespaceFlags,
		AppNamespace:   o.AppGroupFlags.AppNamespace,
	}
	deployOpts.FileFlags = cmdtools.FileFlags{
		Files: []string{app.Path},
	}
	deployOpts.DiffFlags = o.AppFlags.DiffFlags
	deployOpts.ResourceFilterFlags = o.AppFlags.ResourceFilterFlags
	deployOpts.ApplyFlags = o.AppFlags.ApplyFlags
	deployOpts.DeployFlags = o.AppFlags.DeployFlags

	deployOpts.LabelFlags = o.AppFlags.LabelFlags
	deployOpts.LabelFlags.Labels = append(
		deployOpts.LabelFlags.Labels,
		fmt.Sprintf("%s=%s", appGroupAnnKey, o.AppGroupFlags.Name))

	return deployOpts.Run()
}

func (o *DeployOptions) deleteApp(name string) error {
	o.ui.PrintLinef("--- deleting app '%s' (namespace: %s)",
		name, o.appNamespace())

	deleteOpts := cmdapp.NewDeleteOptions(o.ui, o.depsFactory, o.logger)
	deleteOpts.AppFlags = cmdapp.Flags{
		Name:           name,
		NamespaceFlags: o.AppGroupFlags.NamespaceFlags,
		AppNamespace:   o.AppGroupFlags.AppNamespace,
	}
	deleteOpts.DiffFlags = o.AppFlags.DiffFlags
	deleteOpts.ApplyFlags = o.AppFlags.DeleteApplyFlags

	return deleteOpts.Run()
}

func (o *DeployOptions) appNamespace() string {
	if o.AppGroupFlags.AppNamespace != "" {
		return o.AppGroupFlags.AppNamespace
	}
	return o.AppGroupFlags.NamespaceFlags.Name
}
