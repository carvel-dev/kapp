package app

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
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
	ResourceTypesFlags  ResourceTypesFlags
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
	o.ResourceTypesFlags.Set(cmd)
	return cmd
}

func (o *DeleteOptions) Run() error {
	failingAPIServicesPolicy := o.ResourceTypesFlags.FailingAPIServicePolicy()

	app, supportObjs, err := AppFactory(o.depsFactory, o.AppFlags, o.ResourceTypesFlags, o.logger)
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

	usedGVs, err := app.UsedGVs()
	if err != nil {
		return err
	}

	failingAPIServicesPolicy.MarkRequiredGVs(usedGVs)

	existingResources, fullyDeleteApp, err := o.existingResources(app, supportObjs)
	if err != nil {
		return err
	}

	clusterChangeSet, clusterChangesGraph, err := o.calculateAndPresentChanges(existingResources, supportObjs)
	if err != nil {
		return err
	}

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

func (o *DeleteOptions) existingResources(app ctlapp.App,
	supportObjs AppFactorySupportObjs) ([]ctlres.Resource, bool, error) {

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return nil, false, err
	}

	existingResources, err := supportObjs.IdentifiedResources.List(labelSelector)
	if err != nil {
		return nil, false, err
	}

	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return nil, false, err
	}

	fullyDeleteApp := true
	applicableExistingResources := resourceFilter.Apply(existingResources)

	if len(applicableExistingResources) != len(existingResources) {
		fullyDeleteApp = false
		o.ui.PrintLinef("App '%s' (namespace: %s) will not be fully deleted "+
			"because some resources are excluded by filters",
			app.Name(), o.AppFlags.NamespaceFlags.Name)
	}

	existingResources = applicableExistingResources

	o.changeIgnored(existingResources)

	return existingResources, fullyDeleteApp, nil
}

func (o *DeleteOptions) calculateAndPresentChanges(existingResources []ctlres.Resource,
	supportObjs AppFactorySupportObjs) (ctlcap.ClusterChangeSet, *ctldgraph.ChangeGraph, error) {

	var clusterChangeSet ctlcap.ClusterChangeSet

	{ // Figure out changes for X existing resources -> 0 new resources
		changeFactory := ctldiff.NewChangeFactory(nil, nil)
		changeSetFactory := ctldiff.NewChangeSetFactory(o.DiffFlags.ChangeSetOpts, changeFactory)

		changes, err := changeSetFactory.New(existingResources, nil).Calculate()
		if err != nil {
			return ctlcap.ClusterChangeSet{}, nil, err
		}

		{ // Build cluster changes based on diff changes
			msgsUI := cmdcore.NewDedupingMessagesUI(cmdcore.NewPlainMessagesUI(o.ui))

			convergedResFactory := ctlcap.NewConvergedResourceFactory(ctlcap.ConvergedResourceFactoryOpts{
				IgnoreFailingAPIServices: o.ResourceTypesFlags.IgnoreFailingAPIServices,
			})

			clusterChangeFactory := ctlcap.NewClusterChangeFactory(
				o.ApplyFlags.ClusterChangeOpts, supportObjs.IdentifiedResources,
				changeFactory, changeSetFactory, convergedResFactory, msgsUI)

			clusterChangeSet = ctlcap.NewClusterChangeSet(
				changes, o.ApplyFlags.ClusterChangeSetOpts, clusterChangeFactory, msgsUI)
		}
	}

	clusterChanges, clusterChangesGraph, err := clusterChangeSet.Calculate()
	if err != nil {
		return ctlcap.ClusterChangeSet{}, nil, err
	}

	{ // Present cluster changes in UI
		changeViews := ctlcap.ClusterChangesAsChangeViews(clusterChanges)
		changeSetView := ctlcap.NewChangeSetView(changeViews, o.DiffFlags.ChangeSetViewOpts)
		changeSetView.Print(o.ui)
	}

	return clusterChangeSet, clusterChangesGraph, nil
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
