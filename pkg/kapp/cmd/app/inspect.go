package app

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	"github.com/spf13/cobra"
)

type InspectOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags            AppFlags
	ResourceFilterFlags cmdtools.ResourceFilterFlags

	Raw    bool
	Status bool
	Tree   bool
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
		Annotations: map[string]string{
			cmdcore.AppHelpGroup.Key: cmdcore.AppHelpGroup.Value,
		},
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.ResourceFilterFlags.Set(cmd)
	cmd.Flags().BoolVar(&o.Raw, "raw", false, "Output raw YAML resource content")
	cmd.Flags().BoolVar(&o.Status, "status", false, "Output status content")
	cmd.Flags().BoolVarP(&o.Tree, "tree", "t", false, "Tree view")
	return cmd
}

func (o *InspectOptions) Run() error {
	app, _, identifiedResources, err := AppFactory(o.depsFactory, o.AppFlags)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	resources, err := identifiedResources.List(labelSelector)
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
