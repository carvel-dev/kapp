package app

import (
	"fmt"
	"sort"

	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type DeployOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags            AppFlags
	FileFlags           cmdtools.FileFlags
	DiffFlags           cmdtools.DiffFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	ApplyFlags          ApplyFlags
	DeployFlags         DeployFlags
	LabelFlags          LabelFlags
}

func NewDeployOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *DeployOptions {
	return &DeployOptions{ui: ui, depsFactory: depsFactory}
}

func NewDeployCmd(o *DeployOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"d", "dep"},
		Short:   "Deploy app",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.FileFlags.Set(cmd)
	o.DiffFlags.SetWithPrefix("diff", cmd)
	o.ResourceFilterFlags.Set(cmd)
	o.ApplyFlags.SetWithDefaults(ApplyFlagsDeployDefaults, cmd)
	o.DeployFlags.Set(cmd)
	o.LabelFlags.Set(cmd)
	return cmd
}

func (o *DeployOptions) Run() error {
	app, coreClient, dynamicClient, err := appFactory(o.depsFactory, o.AppFlags)
	if err != nil {
		return err
	}

	appLabels, err := o.LabelFlags.AsMap()
	if err != nil {
		return err
	}

	err = app.CreateOrUpdate(appLabels)
	if err != nil {
		return err
	}

	newResources, err := o.newResources()
	if err != nil {
		return err
	}

	newResources, conf, err := ctlconf.NewConfFromResourcesWithDefaults(newResources)
	if err != nil {
		return err
	}

	prep := ctlapp.NewPreparation(coreClient, dynamicClient)

	o.DeployFlags.PrepareResourcesOpts.DefaultNamespace = o.AppFlags.NamespaceFlags.Name

	newResources, err = prep.PrepareResources(newResources, o.DeployFlags.PrepareResourcesOpts)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	identifiedResources := ctlres.NewIdentifiedResources(coreClient, dynamicClient)
	labeledResources := ctlres.NewLabeledResources(labelSelector, identifiedResources)

	err = labeledResources.Prepare(newResources, conf.OwnershipLabelMods(), conf.LabelScopingMods(), conf.AdditionalLabels())
	if err != nil {
		return err
	}

	// Grab ns names before they applying filtering
	nsNames := o.nsNames(newResources)

	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return err
	}

	newResources = resourceFilter.Apply(newResources)
	matchingOpts := ctlres.AllAndMatchingOpts{SkipResourceOwnershipCheck: o.DeployFlags.OverrideOwnershipOfExistingResources}

	existingResources, err := labeledResources.AllAndMatching(newResources, matchingOpts)
	if err != nil {
		return err
	}

	if o.DeployFlags.Patch {
		existingResources, err = ctlres.NewUniqueResources(existingResources).Match(newResources)
		if err != nil {
			return err
		}
	} else {
		if len(newResources) == 0 && !o.DeployFlags.AllowEmpty {
			return fmt.Errorf("Trying to apply empty set of resources. Refusing to continue unless --allow-empty is specified.")
		}
	}

	existingResources = resourceFilter.Apply(existingResources)
	changeFactory := ctldiff.NewChangeFactory(conf.RebaseMods())

	changes, err := ctldiff.NewChangeSetWithTemplates(
		existingResources, newResources, conf.TemplateRules(),
		o.DiffFlags.ChangeSetOpts, changeFactory).Calculate()
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

	changeSetView := ctlcap.NewChangeSetView(ctlcap.ClusterChangesAsChangeViews(clusterChanges), o.DiffFlags.ChangeSetViewOpts)
	changeSetView.Print(o.ui)

	// Validate after showing change set to make it easier to see all resources
	err = prep.ValidateResources(newResources, o.DeployFlags.PrepareResourcesOpts)
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

	touch := ctlapp.Touch{
		App:              app,
		Description:      "update: " + changeSetView.Summary(),
		Namespaces:       nsNames,
		IgnoreSuccessErr: true,
	}

	return touch.Do(func() error {
		return clusterChangeSet.Apply(clusterChangesGraph)
	})
}

func (o *DeployOptions) newResources() ([]ctlres.Resource, error) {
	var allResources []ctlres.Resource

	for _, file := range o.FileFlags.Files {
		fileRs, err := ctlres.NewFileResources(file)
		if err != nil {
			return nil, err
		}

		for _, fileRes := range fileRs {
			resources, err := fileRes.Resources()
			if err != nil {
				return nil, err
			}

			allResources = append(allResources, resources...)
		}
	}

	return allResources, nil
}

func (o *DeployOptions) nsNames(resources []ctlres.Resource) []string {
	uniqNames := map[string]struct{}{}
	names := []string{}
	for _, res := range resources {
		ns := res.Namespace()
		if ns == "" {
			ns = "<cluster>"
		}
		if _, found := uniqNames[ns]; !found {
			names = append(names, ns)
			uniqNames[ns] = struct{}{}
		}
	}
	sort.Strings(names)
	return names
}
