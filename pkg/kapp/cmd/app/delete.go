package app

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags   AppFlags
	DiffFlags  cmdtools.DiffFlags
	ApplyFlags ApplyFlags
}

func NewDeleteOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *DeleteOptions {
	return &DeleteOptions{ui: ui, depsFactory: depsFactory}
}

func NewDeleteCmd(o *DeleteOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete app",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.DiffFlags.SetWithPrefix("diff", cmd)
	o.ApplyFlags.SetWithDefaults("", ApplyFlagsDeleteDefaults, cmd)
	return cmd
}

func (o *DeleteOptions) Run() error {
	app, coreClient, dynamicClient, err := appFactory(o.depsFactory, o.AppFlags)
	if err != nil {
		return err
	}

	exists, err := app.Exists()
	if err != nil {
		return err
	}

	if !exists {
		o.ui.PrintLinef("App '%s' (namespace: %s) does not exist",
			app.Name(), o.AppFlags.NamespaceFlags.Name)
		return nil
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	identifiedResources := ctlres.NewIdentifiedResources(coreClient, dynamicClient, []string{o.AppFlags.NamespaceFlags.Name})

	existingResources, err := identifiedResources.List(labelSelector)
	if err != nil {
		return err
	}

	changeFactory := ctldiff.NewChangeFactory(nil)

	changes, err := ctldiff.NewChangeSet(existingResources, nil, o.DiffFlags.ChangeSetOpts, changeFactory).Calculate()
	if err != nil {
		return err
	}

	msgsUI := cmdcore.NewMessagesUI(o.ui)
	clusterChangeFactory := ctlcap.NewClusterChangeFactory(o.ApplyFlags.ClusterChangeOpts, identifiedResources, changeFactory, msgsUI)
	clusterChangeSet := ctlcap.NewClusterChangeSet(changes, o.ApplyFlags.ClusterChangeSetOpts, clusterChangeFactory, msgsUI)

	clusterChanges, clusterChangesGraph, err := clusterChangeSet.Calculate()
	if err != nil {
		return err
	}

	ctlcap.NewChangeSetView(ctlcap.ClusterChangesAsChangeViews(clusterChanges), o.DiffFlags.ChangeSetViewOpts).Print(o.ui)

	if o.DiffFlags.Run {
		return nil
	}

	err = o.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	touch := ctlapp.Touch{App: app, Description: "delete", IgnoreSuccessErr: true}

	return touch.Do(func() error {
		err := clusterChangeSet.Apply(clusterChangesGraph)
		if err != nil {
			return err
		}

		return app.Delete()
	})
}
