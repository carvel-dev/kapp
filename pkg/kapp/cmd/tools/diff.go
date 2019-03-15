package tools

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctldiff "github.com/k14s/kapp/pkg/kapp/diff"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type DiffOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags  FileFlags
	FileFlags2 FileFlags2
	DiffFlags  DiffFlags
}

func NewDiffOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *DiffOptions {
	return &DiffOptions{ui: ui, depsFactory: depsFactory}
}

func NewDiffCmd(o *DiffOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff files against files2",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.FileFlags2.Set(cmd)
	o.DiffFlags.SetWithPrefix("", cmd)
	return cmd
}

func (o *DiffOptions) Run() error {
	newResources, err := o.fileResources(o.FileFlags.Files, o.FileFlags.Recursive)
	if err != nil {
		return err
	}

	existingResources, err := o.fileResources(o.FileFlags2.Files, o.FileFlags2.Recursive)
	if err != nil {
		return err
	}

	changeFactory := ctldiff.NewChangeFactory(nil)

	changes, err := ctldiff.NewChangeSet(existingResources, newResources, o.DiffFlags.ChangeSetOpts, changeFactory).Calculate()
	if err != nil {
		return err
	}

	ctldiff.NewChangeSetView(changes, o.DiffFlags.ChangeSetViewOpts).Print(o.ui)

	return nil
}

func (o *DiffOptions) fileResources(files []string, recursive bool) ([]ctlres.Resource, error) {
	var newResources []ctlres.Resource

	for _, file := range files {
		fileRs, err := ctlres.NewFileResources(file, recursive)
		if err != nil {
			return nil, err
		}

		for _, fileRes := range fileRs {
			resources, err := fileRes.Resources()
			if err != nil {
				return nil, err
			}

			newResources = append(newResources, resources...)
		}
	}

	return newResources, nil
}
