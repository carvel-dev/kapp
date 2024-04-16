// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	cmdtools "carvel.dev/kapp/pkg/kapp/cmd/tools"
	ctldiff "carvel.dev/kapp/pkg/kapp/diff"
	"carvel.dev/kapp/pkg/kapp/logger"
	"carvel.dev/kapp/pkg/kapp/resources"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type InspectOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags            Flags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	ResourceTypesFlags  ResourceTypesFlags

	Raw           bool
	Status        bool
	Tree          bool
	ManagedFields bool
}

func NewInspectOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *InspectOptions {
	return &InspectOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewInspectCmd(o *InspectOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect",
		Aliases: []string{"i", "is", "insp"},
		Short:   "Inspect app",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
		Annotations: map[string]string{
			cmdcore.AppHelpGroup.Key: cmdcore.AppHelpGroup.Value,
		},
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.ResourceFilterFlags.Set(cmd)
	o.ResourceTypesFlags.Set(cmd)
	cmd.Flags().BoolVar(&o.Raw, "raw", false, "Output raw YAML resource content")
	cmd.Flags().BoolVar(&o.Status, "status", false, "Output status content")
	cmd.Flags().BoolVarP(&o.Tree, "tree", "t", false, "Tree view")
	cmd.Flags().BoolVar(&o.ManagedFields, "managed-fields", false, "Keep the metadata.managedFields when printing objects")
	return cmd
}

func (o *InspectOptions) Run() error {
	failingAPIServicesPolicy := o.ResourceTypesFlags.FailingAPIServicePolicy()

	app, supportObjs, err := Factory(o.depsFactory, o.AppFlags, o.ResourceTypesFlags, o.logger)
	if err != nil {
		return err
	}

	usedGVs, err := app.UsedGVs()
	if err != nil {
		return err
	}

	failingAPIServicesPolicy.MarkRequiredGVs(usedGVs)

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	meta, err := app.Meta()
	if err != nil {
		return err
	}

	resources, err := supportObjs.IdentifiedResources.List(labelSelector, nil, resources.IdentifiedResourcesListOpts{
		ResourceNamespaces: meta.LastChange.Namespaces})
	if err != nil {
		return err
	}

	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return err
	}

	resources = resourceFilter.Apply(resources)
	source := fmt.Sprintf("app '%s'", app.Name())

	switch {
	case o.Raw:
		for _, res := range resources {
			historylessRes, err := ctldiff.NewResourceWithoutHistory(res, nil).Resource()
			if err != nil {
				return err
			}
			resManagedFields, err := ctlres.NewResourceWithManagedFields(historylessRes, o.ManagedFields).Resource()
			if err != nil {
				return err
			}

			resBs, err := resManagedFields.AsYAMLBytes()
			if err != nil {
				return err
			}

			o.ui.PrintBlock(append([]byte("---\n"), resBs...))
		}

	case o.Status:
		InspectStatusView{Source: source, Resources: resources}.Print(o.ui)

	default:
		if o.Tree {
			cmdtools.InspectTreeView{Source: source, Resources: resources, Sort: true}.Print(o.ui)
		} else {
			cmdtools.InspectView{Source: source, Resources: resources, Sort: true}.Print(o.ui)
		}
	}

	return nil
}
