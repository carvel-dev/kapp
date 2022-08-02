// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	cmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/logger"
)

type LabelOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags Flags
}

func NewLabelOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *LabelOptions {
	return &LabelOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewLabelCmd(o *LabelOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Print specified app label",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppSupportHelpGroup.Key: cmdcore.AppSupportHelpGroup.Value,
		},
	}
	o.AppFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *LabelOptions) Run() error {
	app, _, err := Factory(o.depsFactory, o.AppFlags, ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	o.ui.PrintBlock([]byte(labelSelector.String()))

	return nil
}
