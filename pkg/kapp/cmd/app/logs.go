// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	ctllogs "carvel.dev/kapp/pkg/kapp/logs"
	"carvel.dev/kapp/pkg/kapp/matcher"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

type LogsOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags  Flags
	LogsFlags LogsFlags
}

func NewLogsOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *LogsOptions {
	return &LogsOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewLogsCmd(o *LogsOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"l"},
		Short:   "Print app's Pod logs",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppHelpGroup.Key: cmdcore.AppHelpGroup.Value,
		},
		Example: `
  # Follow all pod logs in app 'app1'
  kapp logs -a app1 -f

  # Show logs from pods that start with 'web'
  kapp logs -a app1 -f -m web%`,
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.LogsFlags.Set(cmd)
	return cmd
}

func (o *LogsOptions) Run() error {
	logOpts, err := o.LogsFlags.PodLogOpts()
	if err != nil {
		return err
	}

	app, supportObjs, err := Factory(o.depsFactory, o.AppFlags, ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	podWatcher := ctlres.FilteringPodWatcher{
		func(pod *corev1.Pod) bool {
			if len(o.LogsFlags.PodName) > 0 {
				return matcher.NewStringMatcher(o.LogsFlags.PodName).Matches(pod.Name)
			}
			return true
		},
		supportObjs.IdentifiedResources.PodResources(labelSelector, nil),
	}

	contFilter := func(_ corev1.Pod) []string {
		return o.LogsFlags.ContainerNames
	}

	logsView := ctllogs.NewView(logOpts, podWatcher, contFilter, supportObjs.CoreClient, o.ui)

	return logsView.Show(make(chan struct{}))
}
