// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	ctldiffui "github.com/k14s/kapp/pkg/kapp/diffui"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags            Flags
	DiffFlags           cmdtools.DiffFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	ApplyFlags          ApplyFlags
	ResourceTypesFlags  ResourceTypesFlags
}

type changesSummary struct {
	HasNoChanges   bool
	SkippedChanges bool
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

	app, supportObjs, err := Factory(o.depsFactory, o.AppFlags, o.ResourceTypesFlags, o.logger)
	if err != nil {
		return err
	}

	exists, notExistsMsg, err := app.Exists()
	if err != nil {
		return err
	}

	if !exists {
		o.ui.PrintLinef("%s", notExistsMsg)
		return nil
	}

	usedGVs, err := app.UsedGVs()
	if err != nil {
		return err
	}

	failingAPIServicesPolicy.MarkRequiredGVs(usedGVs)

	existingResources, shouldFullyDeleteApp, err := o.existingResources(app, supportObjs)
	if err != nil {
		return err
	}

	_, conf, err := ctlconf.NewConfFromResourcesWithDefaults(nil)
	if err != nil {
		return err
	}

	clusterChangeSet, clusterChangesGraph, changesSummary, err :=
		o.calculateAndPresentChanges(existingResources, conf, supportObjs)
	if err != nil {
		if o.DiffFlags.UI && clusterChangesGraph != nil {
			return o.presentDiffUI(clusterChangesGraph)
		}
		return err
	}

	if changesSummary.SkippedChanges {
		shouldFullyDeleteApp = false
	}

	if !shouldFullyDeleteApp {
		o.ui.PrintLinef("App '%s' (namespace: %s) will not be fully deleted "+
			"because some resources are excluded by filters",
			app.Name(), o.AppFlags.NamespaceFlags.Name)
	}

	if o.DiffFlags.UI {
		return o.presentDiffUI(clusterChangesGraph)
	}

	if o.DiffFlags.Run {
		if o.DiffFlags.ExitStatus {
			return DeployDiffExitStatus{changesSummary.HasNoChanges}
		}
		return nil
	}

	err = o.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	touch := ctlapp.Touch{App: app, Description: "delete", IgnoreSuccessErr: true}

	err = touch.Do(func() error {
		err := clusterChangeSet.Apply(clusterChangesGraph)
		if err != nil {
			return err
		}
		if shouldFullyDeleteApp {
			return app.Delete()
		}
		return nil
	})
	if err != nil {
		return err
	}

	if o.ApplyFlags.ExitStatus {
		return DeployApplyExitStatus{changesSummary.HasNoChanges}
	}
	return nil
}

func (o *DeleteOptions) existingResources(app ctlapp.App,
	supportObjs FactorySupportObjs) ([]ctlres.Resource, bool, error) {

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return nil, false, err
	}

	existingResources, err := supportObjs.IdentifiedResources.List(labelSelector, nil)
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
	}

	existingResources = applicableExistingResources

	o.changeIgnored(existingResources)

	return existingResources, fullyDeleteApp, nil
}

func (o *DeleteOptions) calculateAndPresentChanges(existingResources []ctlres.Resource, conf ctlconf.Conf,
	supportObjs FactorySupportObjs) (ctlcap.ClusterChangeSet, *ctldgraph.ChangeGraph, changesSummary, error) {

	var (
		clusterChangeSet ctlcap.ClusterChangeSet
		skippedChanges   bool
	)

	{ // Figure out changes for X existing resources -> 0 new resources
		changeFactory := ctldiff.NewChangeFactory(nil, nil)
		changeSetFactory := ctldiff.NewChangeSetFactory(o.DiffFlags.ChangeSetOpts, changeFactory)

		changes, err := changeSetFactory.New(existingResources, nil).Calculate()
		if err != nil {
			return ctlcap.ClusterChangeSet{}, nil, changesSummary{}, err
		}

		diffFilter, err := o.DiffFlags.DiffFilter()
		if err != nil {
			return ctlcap.ClusterChangeSet{}, nil, changesSummary{}, err
		}

		appliedChanges := diffFilter.Apply(changes)

		if len(changes) != len(appliedChanges) {
			// setting it to true when changes are filtered by diffFilter
			skippedChanges = true
		}

		{ // Build cluster changes based on diff changes
			msgsUI := cmdcore.NewDedupingMessagesUI(cmdcore.NewPlainMessagesUI(o.ui))

			convergedResFactory := ctlcap.NewConvergedResourceFactory(conf.WaitRules(), ctlcap.ConvergedResourceFactoryOpts{
				IgnoreFailingAPIServices: o.ResourceTypesFlags.IgnoreFailingAPIServices,
			})

			clusterChangeFactory := ctlcap.NewClusterChangeFactory(
				o.ApplyFlags.ClusterChangeOpts, supportObjs.IdentifiedResources,
				changeFactory, changeSetFactory, convergedResFactory, msgsUI)

			clusterChangeSet = ctlcap.NewClusterChangeSet(
				appliedChanges, o.ApplyFlags.ClusterChangeSetOpts, clusterChangeFactory,
				conf.ChangeGroupBindings(), conf.ChangeRuleBindings(), msgsUI, o.logger)
		}
	}

	clusterChanges, clusterChangesGraph, err := clusterChangeSet.Calculate()
	if err != nil {
		return ctlcap.ClusterChangeSet{}, nil, changesSummary{}, err
	}

	{ // Present cluster changes in UI
		changeViews := ctlcap.ClusterChangesAsChangeViews(clusterChanges)
		changeSetView := ctlcap.NewChangeSetView(
			changeViews, conf.DiffMaskRules(), o.DiffFlags.ChangeSetViewOpts)
		changeSetView.Print(o.ui)
	}

	return clusterChangeSet, clusterChangesGraph, changesSummary{HasNoChanges: len(clusterChanges) == 0, SkippedChanges: skippedChanges}, nil
}

const (
	ownedForDeletionAnnKey = "kapp.k14s.io/owned-for-deletion" // valid values: ''
)

func (o *DeleteOptions) changeIgnored(resources []ctlres.Resource) {
	// Good example for use of this annotation is PVCs created by StatefulSet
	// (PVCs do not get deleted when StatefulSet is deleted:
	// https://github.com/vmware-tanzu/carvel-kapp/issues/36)
	for _, res := range resources {
		if _, found := res.Annotations()[ownedForDeletionAnnKey]; found {
			res.MarkTransient(false)
		}
	}
}

func (o *DeleteOptions) presentDiffUI(graph *ctldgraph.ChangeGraph) error {
	opts := ctldiffui.ServerOpts{
		DiffDataFunc: func() *ctldgraph.ChangeGraph { return graph },
	}
	return ctldiffui.NewServer(opts, o.ui).Run()
}
