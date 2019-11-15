package app

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctllogs "github.com/k14s/kapp/pkg/kapp/logs"
	"github.com/k14s/kapp/pkg/kapp/matcher"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

type LogsOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags  AppFlags
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

	app, supportObjs, err := AppFactory(o.depsFactory, o.AppFlags, ResourceTypesFlags{}, o.logger)
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
		supportObjs.IdentifiedResources.PodResources(labelSelector),
	}

	logsView := ctllogs.NewView(logOpts, podWatcher, supportObjs.CoreClient, o.ui)

	return logsView.Show(make(chan struct{}))
}
