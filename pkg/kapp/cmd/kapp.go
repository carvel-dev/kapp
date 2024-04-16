// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"io"

	cmdapp "carvel.dev/kapp/pkg/kapp/cmd/app"
	cmdac "carvel.dev/kapp/pkg/kapp/cmd/appchange"
	cmdag "carvel.dev/kapp/pkg/kapp/cmd/appgroup"
	cmdcm "carvel.dev/kapp/pkg/kapp/cmd/configmap"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	cmdsa "carvel.dev/kapp/pkg/kapp/cmd/serviceaccount"
	cmdtools "carvel.dev/kapp/pkg/kapp/cmd/tools"
	"carvel.dev/kapp/pkg/kapp/crdupgradesafety"
	"carvel.dev/kapp/pkg/kapp/logger"
	"carvel.dev/kapp/pkg/kapp/permissions"
	"carvel.dev/kapp/pkg/kapp/preflight"
	"carvel.dev/kapp/pkg/kapp/version"
	"github.com/cppforlife/cobrautil"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type KappOptions struct {
	ui            *ui.ConfUI
	logger        *logger.UILogger
	configFactory cmdcore.ConfigFactory
	depsFactory   cmdcore.DepsFactory

	UIFlags         UIFlags
	LoggerFlags     LoggerFlags
	KubeAPIFlags    cmdcore.KubeAPIFlags
	KubeconfigFlags cmdcore.KubeconfigFlags
	WarningFlags    WarningFlags
	ProfilingFlags  ProfilingFlags

	PreflightChecks *preflight.Registry
}

func NewKappOptions(ui *ui.ConfUI, configFactory cmdcore.ConfigFactory,
	depsFactory cmdcore.DepsFactory, preflights *preflight.Registry) *KappOptions {

	return &KappOptions{ui: ui, logger: logger.NewUILogger(ui),
		configFactory: configFactory, depsFactory: depsFactory, PreflightChecks: preflights}
}

func NewDefaultKappCmd(ui *ui.ConfUI) *cobra.Command {
	configFactory := cmdcore.NewConfigFactoryImpl()
	depsFactory := cmdcore.NewDepsFactoryImpl(configFactory, ui)
	preflights := defaultKappPreflightRegistry(depsFactory)
	options := NewKappOptions(ui, configFactory, depsFactory, preflights)
	flagsFactory := cmdcore.NewFlagsFactory(configFactory, depsFactory)
	return NewKappCmd(options, flagsFactory)
}

func defaultKappPreflightRegistry(depsFactory cmdcore.DepsFactory) *preflight.Registry {
	registry := preflight.NewRegistry(map[string]preflight.Check{
		"PermissionValidation": permissions.NewPreflight(depsFactory, false),
		"CRDUpgradeSafety":     crdupgradesafety.NewPreflight(depsFactory, false),
	})

	return registry
}

func NewKappCmd(o *KappOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kapp",
		Short: "kapp helps to manage applications on your Kubernetes cluster",

		RunE: cobrautil.ShowHelp,

		// Affects children as well
		SilenceErrors: true,
		SilenceUsage:  true,

		// Disable docs header
		DisableAutoGenTag: true,
		Version:           version.Version,

		PersistentPostRunE: func(*cobra.Command, []string) error {
			return o.ProfilingFlags.flushProfiling()
		},

		// TODO bash completion
	}

	cmd.SetOut(uiBlockWriter{o.ui}) // setting output for cmd.Help()

	cmd.SetUsageTemplate(cobrautil.HelpSectionsUsageTemplate([]cobrautil.HelpSection{
		cmdcore.AppHelpGroup,
		cmdcore.AppSupportHelpGroup,
		cmdcore.MiscHelpGroup,
		cmdcore.RestOfCommandsHelpGroup,
	}))

	o.UIFlags.Set(cmd, flagsFactory)
	o.LoggerFlags.Set(cmd, flagsFactory)
	o.KubeAPIFlags.Set(cmd, flagsFactory)
	o.KubeconfigFlags.Set(cmd, flagsFactory)
	o.WarningFlags.Set(cmd, flagsFactory)
	o.ProfilingFlags.Set(cmd, flagsFactory)

	o.configFactory.ConfigurePathResolver(o.KubeconfigFlags.Path.Value)
	o.configFactory.ConfigureContextResolver(o.KubeconfigFlags.Context.Value)
	o.configFactory.ConfigureYAMLResolver(o.KubeconfigFlags.YAML.Value)

	cmd.AddCommand(NewVersionCmd(NewVersionOptions(o.ui), flagsFactory))

	cmd.AddCommand(cmdapp.NewListCmd(cmdapp.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewInspectCmd(cmdapp.NewInspectOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewDeployCmd(cmdapp.NewDeployOptions(o.ui, o.depsFactory, o.logger, o.PreflightChecks), flagsFactory))
	cmd.AddCommand(cmdapp.NewDeployConfigCmd(cmdapp.NewDeployConfigOptions(o.ui, o.depsFactory), flagsFactory))
	cmd.AddCommand(cmdapp.NewDeleteCmd(cmdapp.NewDeleteOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewRenameCmd(cmdapp.NewRenameOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewLogsCmd(cmdapp.NewLogsOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmdapp.NewLabelCmd(cmdapp.NewLabelOptions(o.ui, o.depsFactory, o.logger), flagsFactory))

	agCmd := cmdag.NewCmd()
	agCmd.AddCommand(cmdag.NewDeployCmd(cmdag.NewDeployOptions(o.ui, o.depsFactory, o.logger, o.PreflightChecks), flagsFactory))
	agCmd.AddCommand(cmdag.NewDeleteCmd(cmdag.NewDeleteOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(agCmd)

	cmCmd := cmdcm.NewCmd()
	cmCmd.AddCommand(cmdcm.NewListCmd(cmdcm.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(cmCmd)

	acCmd := cmdac.NewCmd()
	acCmd.AddCommand(cmdac.NewListCmd(cmdac.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	acCmd.AddCommand(cmdac.NewGCCmd(cmdac.NewGCOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(acCmd)

	saCmd := cmdsa.NewCmd()
	saCmd.AddCommand(cmdsa.NewListCmd(cmdsa.NewListOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(saCmd)

	appCmd := cmdtools.NewCmd()
	appCmd.AddCommand(cmdtools.NewInspectCmd(cmdtools.NewInspectOptions(o.ui, o.depsFactory), flagsFactory))
	appCmd.AddCommand(cmdtools.NewDiffCmd(cmdtools.NewDiffOptions(o.ui, o.depsFactory), flagsFactory))
	appCmd.AddCommand(cmdtools.NewListLabelsCmd(cmdtools.NewListLabelsOptions(o.ui, o.depsFactory, o.logger), flagsFactory))
	cmd.AddCommand(appCmd)

	finishDebugLog := func(cmd *cobra.Command) {
		origRunE := cmd.RunE
		if origRunE != nil {
			cmd.RunE = func(cmd2 *cobra.Command, args []string) error {
				defer o.logger.DebugFunc("CommandRun").Finish()
				return origRunE(cmd2, args)
			}
		}
	}

	configureGlobal := cobrautil.WrapRunEForCmd(func(*cobra.Command, []string) error {
		o.UIFlags.ConfigureUI(o.ui)
		o.LoggerFlags.Configure(o.logger)
		o.KubeAPIFlags.Configure(o.configFactory)
		o.WarningFlags.Configure(o.depsFactory)
		o.ProfilingFlags.initProfiling()
		return nil
	})

	configureTTYFlag := func(cmd *cobra.Command) {
		var forceTTY bool
		// Default tty to true for deploy and delete commands: https://github.com/carvel-dev/kapp/issues/28
		_, defaultTTY := cmd.Annotations[cmdapp.TTYByDefaultKey]

		cmd.Flags().BoolVar(&forceTTY, "tty", defaultTTY, "Force TTY-like output")
		cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(func(*cobra.Command, []string) error {
			o.ui.EnableTTY(forceTTY)
			return nil
		}))
	}

	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureLeafCmds(cobrautil.DisallowExtraArgs))

	// Completion command have to be added after the cobrautil.DisallowExtraArgs.
	// This due to the ReconfigureLeafCmds that we do not want to have enforced for the completion
	// This configurations forces all nodes to do not accept extra args, but the completion requires 1 extra arg
	cmd.AddCommand(NewCmdCompletion())

	// Last one runs first
	cobrautil.VisitCommands(cmd, finishDebugLog, cobrautil.ReconfigureCmdWithSubcmd, configureGlobal,
		cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd), cobrautil.ReconfigureLeafCmds(configureTTYFlag))
	return cmd
}

type uiBlockWriter struct {
	ui ui.UI
}

var _ io.Writer = uiBlockWriter{}

func (w uiBlockWriter) Write(p []byte) (n int, err error) {
	w.ui.PrintBlock(p)
	return len(p), nil
}
