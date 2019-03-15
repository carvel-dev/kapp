package app

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type InspectOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags            AppFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags
	Raw                 bool
	Tree                bool
}

func NewInspectOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *InspectOptions {
	return &InspectOptions{ui: ui, depsFactory: depsFactory}
}

func NewInspectCmd(o *InspectOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect",
		Aliases: []string{"i", "is", "insp"},
		Short:   "Inspect app",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.ResourceFilterFlags.Set(cmd)
	cmd.Flags().BoolVar(&o.Raw, "raw", false, "Output raw YAML resource content")
	cmd.Flags().BoolVarP(&o.Tree, "tree", "t", false, "Tree view")
	return cmd
}

func (o *InspectOptions) Run() error {
	app, coreClient, dynamicClient, err := appFactory(o.depsFactory, o.AppFlags)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	identifiedResources := ctlres.NewIdentifiedResources(coreClient, dynamicClient)

	resources, err := identifiedResources.List(labelSelector)
	if err != nil {
		return err
	}

	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return err
	}

	resources = resourceFilter.Apply(resources)

	if o.Raw {
		for _, res := range resources {
			historylessRes, err := ctldiff.NewResourceWithHistory(res, nil).HistorylessResource()
			if err != nil {
				return err
			}

			resBs, err := historylessRes.AsYAMLBytes()
			if err != nil {
				return err
			}

			o.ui.PrintBlock(append([]byte("---\n"), resBs...))
		}
	} else {
		if o.Tree {
			cmdtools.InspectTreeView{
				Source:    fmt.Sprintf("app '%s'", app.Name()),
				Resources: resources,
				Sort:      true,
			}.Print(o.ui)
		} else {
			cmdtools.InspectView{
				Source:    fmt.Sprintf("app '%s'", app.Name()),
				Resources: resources,
				Sort:      true,
			}.Print(o.ui)
		}
	}

	return nil
}
