// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"sort"
	"strings"

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
	ctllogs "github.com/k14s/kapp/pkg/kapp/logs"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type DeployOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags            Flags
	FileFlags           cmdtools.FileFlags
	DiffFlags           cmdtools.DiffFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	ApplyFlags          ApplyFlags
	DeployFlags         DeployFlags
	ResourceTypesFlags  ResourceTypesFlags
	LabelFlags          LabelFlags
}

func NewDeployOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *DeployOptions {
	return &DeployOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewDeployCmd(o *DeployOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"d", "dep"},
		Short:   "Deploy app",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppHelpGroup.Key: cmdcore.AppHelpGroup.Value,
		},
		Example: `
  # Deploy app 'app1' based on config files in config/
  kapp deploy -a app1 -f config/

  # Deploy app 'app1' while showing full text diff
  kapp deploy -a app1 -f config/ --diff-changes

  # Deploy app 'app1' based on remote file
  kapp deploy -a app1 \
    -f https://github.com/...download/v0.6.0/crds.yaml \
    -f https://github.com/...download/v0.6.0/release.yaml`,
	}

	setDeployCmdFlags(cmd)

	o.AppFlags.Set(cmd, flagsFactory)
	o.FileFlags.Set(cmd)
	o.DiffFlags.SetWithPrefix("diff", cmd)
	o.ResourceFilterFlags.Set(cmd)
	o.ApplyFlags.SetWithDefaults("", ApplyFlagsDeployDefaults, cmd)
	o.DeployFlags.Set(cmd)
	o.ResourceTypesFlags.Set(cmd)
	o.LabelFlags.Set(cmd)

	return cmd
}

func (o *DeployOptions) Run() error {
	failingAPIServicesPolicy := o.ResourceTypesFlags.FailingAPIServicePolicy()

	app, supportObjs, err := Factory(o.depsFactory, o.AppFlags, o.ResourceTypesFlags, o.logger)
	if err != nil {
		return err
	}

	appLabels, err := o.LabelFlags.AsMap()
	if err != nil {
		return err
	}

	if o.DeployFlags.PrevApp != "" {
		err = app.RenamePrevApp(o.DeployFlags.PrevApp, appLabels)
	} else {
		err = app.CreateOrUpdate(appLabels)
	}

	if err != nil {
		return err
	}

	usedGVs, err := app.UsedGVs()
	if err != nil {
		return err
	}

	failingAPIServicesPolicy.MarkRequiredGVs(usedGVs)

	o.DeployFlags.PrepareResourcesOpts.BeforeModificationFunc = func(rs []ctlres.Resource) []ctlres.Resource {
		failingAPIServicesPolicy.MarkRequiredResources(rs)
		return rs
	}

	o.DeployFlags.PrepareResourcesOpts.DefaultNamespace = o.AppFlags.NamespaceFlags.Name

	prep := ctlapp.NewPreparation(supportObjs.ResourceTypes, o.DeployFlags.PrepareResourcesOpts)

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	labeledResources := ctlres.NewLabeledResources(labelSelector, supportObjs.IdentifiedResources, o.logger)

	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return err
	}

	newResources, conf, nsNames, newGKs, err := o.newResources(prep, labeledResources, resourceFilter)
	if err != nil {
		return err
	}

	usedGKs, err := o.newAndUsedGKs(newGKs, app)
	if err != nil {
		return err
	}

	if o.DeployFlags.Logs {
		usedGKs = append(usedGKs, schema.GroupKind{Kind: "Pod"})
	}

	existingResources, existingPodRs, err := o.existingResources(
		newResources, labeledResources, resourceFilter, supportObjs.Apps, usedGKs)
	if err != nil {
		return err
	}

	clusterChangeSet, clusterChangesGraph, hasNoChanges, changeSummary, err :=
		o.calculateAndPresentChanges(existingResources, newResources, conf, supportObjs)
	if err != nil {
		if o.DiffFlags.UI && clusterChangesGraph != nil {
			return o.presentDiffUI(clusterChangesGraph)
		}
		return err
	}

	// Validate new resources _after_ presenting changes to make it easier to see big picture
	err = prep.ValidateResources(newResources)
	if err != nil {
		return err
	}

	if o.DiffFlags.UI {
		return o.presentDiffUI(clusterChangesGraph)
	}

	if o.DiffFlags.Run || hasNoChanges {
		if o.DiffFlags.Run && o.DiffFlags.ExitStatus {
			return DeployDiffExitStatus{hasNoChanges}
		}
		if o.ApplyFlags.ExitStatus {
			return DeployApplyExitStatus{hasNoChanges}
		}
		return nil
	}

	err = o.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	// Track newly added GVs and GKs
	err = app.UpdateUsedGVsAndGKs(failingAPIServicesPolicy.GVs(newResources, existingResources),
		NewUsedGKsScope(append(newResources, existingResources...)).GKs())
	if err != nil {
		return err
	}

	if o.DeployFlags.Logs {
		cancelLogsCh := make(chan struct{})
		defer func() { close(cancelLogsCh) }()
		go o.showLogs(supportObjs.CoreClient, supportObjs.IdentifiedResources, existingPodRs, labelSelector, cancelLogsCh)
	}

	defer func() {
		_, numDeleted, _ := app.GCChanges(o.DeployFlags.AppChangesMaxToKeep, nil)
		if numDeleted > 0 {
			o.ui.PrintLinef("Deleted %d older app changes", numDeleted)
		}
	}()

	touch := ctlapp.Touch{
		App:              app,
		Description:      "update: " + changeSummary,
		Namespaces:       nsNames,
		IgnoreSuccessErr: true,
	}

	err = touch.Do(func() error {
		err := clusterChangeSet.Apply(clusterChangesGraph)
		if err != nil {
			return err
		}

		// Remove unused GVs and GKs
		return app.UpdateUsedGVsAndGKs(failingAPIServicesPolicy.GVs(newResources, nil),
			NewUsedGKsScope(newResources).GKs())
	})
	if err != nil {
		return err
	}

	if o.ApplyFlags.ExitStatus {
		return DeployApplyExitStatus{hasNoChanges}
	}
	return nil
}

func (o *DeployOptions) newAndUsedGKs(newGKs []schema.GroupKind, app ctlapp.App) ([]schema.GroupKind, error) {
	if o.ResourceTypesFlags.DisableGKScoping {
		return []schema.GroupKind{}, nil
	}

	gksByGK := map[schema.GroupKind]struct{}{}
	var uniqGKs []schema.GroupKind

	usedGKs, err := app.UsedGKs()
	if err != nil {
		return nil, err
	}

	// Handle existing apps without cached GKs
	// These apps can cache and scope to GKs in subsequent deploys
	if usedGKs == nil {
		return []schema.GroupKind{}, nil
	}

	for _, gk := range *usedGKs {
		if _, found := gksByGK[gk]; !found {
			gksByGK[gk] = struct{}{}
			uniqGKs = append(uniqGKs, gk)
		}
	}

	for _, gk := range newGKs {
		if _, found := gksByGK[gk]; !found {
			gksByGK[gk] = struct{}{}
			uniqGKs = append(uniqGKs, gk)
		}
	}

	return uniqGKs, nil
}

func (o *DeployOptions) newResources(
	prep ctlapp.Preparation, labeledResources *ctlres.LabeledResources,
	resourceFilter ctlres.ResourceFilter) ([]ctlres.Resource, ctlconf.Conf, []string, []schema.GroupKind, error) {

	newResources, err := o.newResourcesFromFiles()
	if err != nil {
		return nil, ctlconf.Conf{}, nil, nil, err
	}

	newResources, conf, err := ctlconf.NewConfFromResourcesWithDefaults(newResources)
	if err != nil {
		return nil, ctlconf.Conf{}, nil, nil, err
	}

	newResources, err = prep.PrepareResources(newResources)
	if err != nil {
		return nil, ctlconf.Conf{}, nil, nil, err
	}

	err = labeledResources.Prepare(newResources, conf.OwnershipLabelMods(),
		conf.LabelScopingMods(), conf.AdditionalLabels())
	if err != nil {
		return nil, ctlconf.Conf{}, nil, nil, err
	}

	newGKs := NewUsedGKsScope(newResources).GKs()

	// Grab ns names before resource filtering is applied
	nsNames := o.nsNames(newResources)

	return resourceFilter.Apply(newResources), conf, nsNames, newGKs, nil
}

func (o *DeployOptions) newResourcesFromFiles() ([]ctlres.Resource, error) {
	var allResources []ctlres.Resource

	if len(o.FileFlags.Files) == 0 {
		return nil, fmt.Errorf("Expected at least one --file (-f) specified with a file or directory path")
	}
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

func (o *DeployOptions) existingResources(newResources []ctlres.Resource,
	labeledResources *ctlres.LabeledResources, resourceFilter ctlres.ResourceFilter,
	apps ctlapp.Apps, usedGKs []schema.GroupKind) ([]ctlres.Resource, []ctlres.Resource, error) {

	labelErrorResolutionFunc := func(key string, val string) string {
		items, _ := apps.List(nil)
		for _, item := range items {
			meta, _ := item.Meta()
			if meta.LabelKey == key && meta.LabelValue == val {
				return fmt.Sprintf("different %s (label '%s=%s')", item.Description(), key, val)
			}
		}
		return ""
	}

	matchingOpts := ctlres.AllAndMatchingOpts{
		ExistingNonLabeledResourcesCheck:            o.DeployFlags.ExistingNonLabeledResourcesCheck,
		ExistingNonLabeledResourcesCheckConcurrency: o.DeployFlags.ExistingNonLabeledResourcesCheckConcurrency,
		SkipResourceOwnershipCheck:                  o.DeployFlags.OverrideOwnershipOfExistingResources,

		// Prevent accidently overriding kapp state records
		DisallowedResourcesByLabelKeys: []string{ctlapp.KappIsAppLabelKey},
		LabelErrorResolutionFunc:       labelErrorResolutionFunc,

		//Scope resource searching to UsedGKs
		IdentifiedResourcesListOpts: ctlres.IdentifiedResourcesListOpts{
			GKsScope: usedGKs,
		},
	}

	existingResources, err := labeledResources.AllAndMatching(newResources, matchingOpts)
	if err != nil {
		return nil, nil, err
	}

	if o.DeployFlags.Patch {
		existingResources, err = ctlres.NewUniqueResources(existingResources).Match(newResources)
		if err != nil {
			return nil, nil, err
		}
	} else {
		if len(newResources) == 0 && !o.DeployFlags.AllowEmpty {
			return nil, nil, fmt.Errorf("Trying to apply empty set of resources will result in deletion of resources on cluster. " +
				"Refusing to continue unless --dangerous-allow-empty-list-of-resources is specified.")
		}
	}

	return resourceFilter.Apply(existingResources), o.existingPodResources(existingResources), nil
}

func (o *DeployOptions) calculateAndPresentChanges(existingResources,
	newResources []ctlres.Resource, conf ctlconf.Conf, supportObjs FactorySupportObjs) (
	ctlcap.ClusterChangeSet, *ctldgraph.ChangeGraph, bool, string, error) {

	var clusterChangeSet ctlcap.ClusterChangeSet

	{ // Figure out changes for X existing resources -> X new resources
		changeFactory := ctldiff.NewChangeFactory(conf.RebaseMods(), conf.DiffAgainstLastAppliedFieldExclusionMods())
		changeSetFactory := ctldiff.NewChangeSetFactory(o.DiffFlags.ChangeSetOpts, changeFactory)

		changes, err := ctldiff.NewChangeSetWithVersionedRs(
			existingResources, newResources, conf.TemplateRules(),
			o.DiffFlags.ChangeSetOpts, changeFactory).Calculate()
		if err != nil {
			return clusterChangeSet, nil, false, "", err
		}

		diffFilter, err := o.DiffFlags.DiffFilter()
		if err != nil {
			return clusterChangeSet, nil, false, "", err
		}

		changes = diffFilter.Apply(changes)

		msgsUI := cmdcore.NewDedupingMessagesUI(cmdcore.NewPlainMessagesUI(o.ui))

		convergedResFactory := ctlcap.NewConvergedResourceFactory(conf.WaitRules(), ctlcap.ConvergedResourceFactoryOpts{
			IgnoreFailingAPIServices: o.ResourceTypesFlags.IgnoreFailingAPIServices,
		})

		clusterChangeFactory := ctlcap.NewClusterChangeFactory(
			o.ApplyFlags.ClusterChangeOpts, supportObjs.IdentifiedResources,
			changeFactory, changeSetFactory, convergedResFactory, msgsUI, conf.DiffMaskRules())

		clusterChangeSet = ctlcap.NewClusterChangeSet(
			changes, o.ApplyFlags.ClusterChangeSetOpts, clusterChangeFactory,
			conf.ChangeGroupBindings(), conf.ChangeRuleBindings(), msgsUI, o.logger)
	}

	clusterChanges, clusterChangesGraph, err := clusterChangeSet.Calculate()
	if err != nil {
		// Return graph for inspection
		return clusterChangeSet, clusterChangesGraph, false, "", err
	}

	var changesSummary string

	{ // Present cluster changes in UI
		changeViews := ctlcap.ClusterChangesAsChangeViews(clusterChanges)
		changeSetView := ctlcap.NewChangeSetView(
			changeViews, conf.DiffMaskRules(), o.DiffFlags.ChangeSetViewOpts)
		changeSetView.Print(o.ui)
		changesSummary = changeSetView.Summary()
	}

	return clusterChangeSet, clusterChangesGraph, (len(clusterChanges) == 0), changesSummary, err
}

func (o *DeployOptions) existingPodResources(existingResources []ctlres.Resource) []ctlres.Resource {
	var existingPods []ctlres.Resource
	for _, res := range existingResources {
		if ctlresm.NewCoreV1Pod(res) != nil {
			existingPods = append(existingPods, res)
		}
	}
	return existingPods
}

const (
	deployLogsAnnKey              = "kapp.k14s.io/deploy-logs" // valid value is '' (default), for-new, for-existing, for-new-or-existing
	deployLogsAnnDefault          = ""                         // equivalent to for-new
	deployLogsAnnForNew           = "for-new"
	deployLogsAnnForExisting      = "for-existing"
	deployLogsAnnForNewOrExisting = "for-new-or-existing"

	deployLogsContNamesAnnKey = "kapp.k14s.io/deploy-logs-container-names"
)

func (o *DeployOptions) showLogs(
	coreClient kubernetes.Interface, identifiedResources ctlres.IdentifiedResources,
	existingPodRs []ctlres.Resource, labelSelector labels.Selector, cancelCh chan struct{}) {

	existingPodsByUID := map[string]struct{}{}

	for _, res := range existingPodRs {
		existingPodsByUID[res.UID()] = struct{}{}
	}

	podMatcherFunc := func(pod *corev1.Pod) bool {
		if o.DeployFlags.LogsAll {
			return true
		}

		lvl, showLogs := pod.Annotations[deployLogsAnnKey]
		if !showLogs {
			return false
		}

		_, isExistingPod := existingPodsByUID[string(pod.UID)]

		switch lvl {
		case deployLogsAnnDefault, deployLogsAnnForNew:
			return !isExistingPod
		case deployLogsAnnForExisting:
			return isExistingPod
		case deployLogsAnnForNewOrExisting:
			return true
		default:
			return false
		}
	}

	podWatcher := ctlres.FilteringPodWatcher{
		podMatcherFunc,
		identifiedResources.PodResources(labelSelector),
	}

	contFilterFunc := func(pod corev1.Pod) []string {
		ann, found := pod.Annotations[deployLogsContNamesAnnKey]
		if found && ann != "" {
			return strings.Split(ann, ",")
		}
		return nil
	}

	logOpts := ctllogs.PodLogOpts{Follow: true, ContainerTag: true, LinePrefix: "logs"}

	ctllogs.NewView(logOpts, podWatcher, contFilterFunc, coreClient, o.ui).Show(cancelCh)
}

func (o *DeployOptions) nsNames(resources []ctlres.Resource) []string {
	uniqNames := map[string]struct{}{}
	names := []string{}
	for _, res := range resources {
		ns := res.Namespace()
		if ns == "" {
			ns = "(cluster)"
		}
		if _, found := uniqNames[ns]; !found {
			names = append(names, ns)
			uniqNames[ns] = struct{}{}
		}
	}
	sort.Strings(names)
	return names
}

func (o *DeployOptions) presentDiffUI(graph *ctldgraph.ChangeGraph) error {
	opts := ctldiffui.ServerOpts{
		DiffDataFunc: func() *ctldgraph.ChangeGraph { return graph },
	}
	return ctldiffui.NewServer(opts, o.ui).Run()
}
