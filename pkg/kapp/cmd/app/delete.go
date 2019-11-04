package app

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags            AppFlags
	DiffFlags           cmdtools.DiffFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	ApplyFlags          ApplyFlags
}

func NewDeleteOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *DeleteOptions {
	return &DeleteOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewDeleteCmd(o *DeleteOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete app",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppHelpGroup.Key: cmdcore.AppHelpGroup.Value,
		},
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.DiffFlags.SetWithPrefix("diff", cmd)
	o.ResourceFilterFlags.Set(cmd)
	o.ApplyFlags.SetWithDefaults("", ApplyFlagsDeleteDefaults, cmd)
	return cmd
}

func (o *DeleteOptions) Run() error {
	app, _, identifiedResources, err := AppFactory(o.depsFactory, o.AppFlags, o.logger)
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

	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return err
	}

	existingResources, err := identifiedResources.List(labelSelector)
	if err != nil {
		return err
	}

	fullyDeleteApp := true
	applicableExistingResources := resourceFilter.Apply(existingResources)

	if len(applicableExistingResources) != len(existingResources) {
		fullyDeleteApp = false
		o.ui.PrintLinef("App '%s' (namespace: %s) will not be fully deleted because some resources are excluded by filters",
			app.Name(), o.AppFlags.NamespaceFlags.Name)
	}

	existingResources = applicableExistingResources
	changeFactory := ctldiff.NewChangeFactory(nil, nil)

	o.changeIgnored(existingResources)

	changeSetFactory := ctldiff.NewChangeSetFactory(o.DiffFlags.ChangeSetOpts, changeFactory)

	changes, err := changeSetFactory.New(existingResources, nil).Calculate()
	if err != nil {
		return err
	}

	msgsUI := cmdcore.NewDedupingMessagesUI(cmdcore.NewPlainMessagesUI(o.ui))
	clusterChangeFactory := ctlcap.NewClusterChangeFactory(o.ApplyFlags.ClusterChangeOpts, identifiedResources, changeFactory, changeSetFactory, msgsUI)
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

		if fullyDeleteApp {
			return app.Delete()
		}

		return nil
	})
}

const (
	ownedForDeletionAnnKey = "kapp.k14s.io/owned-for-deletion" // valid values: ''
)

func (o *DeleteOptions) changeIgnored(resources []ctlres.Resource) {
	// Good example for use of this annotation is PVCs created by StatefulSet
	// (PVCs do not get deleted when StatefulSet is deleted:
	// https://github.com/k14s/kapp/issues/36)
	for _, res := range resources {
		if _, found := res.Annotations()[ownedForDeletionAnnKey]; found {
			res.MarkTransient(false)
		}
	}
}
