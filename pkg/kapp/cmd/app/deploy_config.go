// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	ctlconf "carvel.dev/kapp/pkg/kapp/config"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type DeployConfigOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
}

func NewDeployConfigOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *DeployConfigOptions {
	return &DeployConfigOptions{ui: ui, depsFactory: depsFactory}
}

func NewDeployConfigCmd(o *DeployConfigOptions, _ cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-config",
		Short: "Show default deploy config",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.MiscHelpGroup.Key: cmdcore.MiscHelpGroup.Value,
		},
	}
	return cmd
}

func (o *DeployConfigOptions) Run() error {
	o.ui.PrintBlock([]byte(ctlconf.NewDefaultConfigString()))

	return nil
}
