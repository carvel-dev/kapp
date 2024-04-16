// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"io/fs"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type InspectOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags           FileFlags
	ResourceFilterFlags ResourceFilterFlags
	Raw                 bool

	FileSystem fs.FS
}

func NewInspectOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *InspectOptions {
	return &InspectOptions{ui: ui, depsFactory: depsFactory}
}

func NewInspectCmd(o *InspectOptions, _ cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect",
		Aliases: []string{"i"},
		Short:   "Inspect resources",
		Long:    "Inspect resources",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.ResourceFilterFlags.Set(cmd)
	cmd.Flags().BoolVar(&o.Raw, "raw", false, "Output raw YAML resource content")
	return cmd
}

func (o *InspectOptions) Run() error {
	// Only supports inspecting files; use kapp -a k=v inspect for cluster inspection
	return o.inspectFiles()
}

func (o *InspectOptions) inspectFiles() error {
	resourceFilter, err := o.ResourceFilterFlags.ResourceFilter()
	if err != nil {
		return err
	}

	for _, file := range o.FileFlags.Files {
		fileRs, err := ctlres.NewFileResources(o.FileSystem, file)
		if err != nil {
			return err
		}

		for _, fileRes := range fileRs {
			resources, err := fileRes.Resources()
			if err != nil {
				return err
			}

			resources = resourceFilter.Apply(resources)

			if o.Raw {
				for _, res := range resources {
					resBs, err := res.AsYAMLBytes()
					if err != nil {
						return err
					}

					o.ui.PrintBlock(append([]byte("---\n"), resBs...))
				}
			} else {
				view := InspectView{
					Source:    fileRes.Description(),
					Resources: resources,
					Sort:      o.FileFlags.Sort,
				}

				view.Print(o.ui)
			}
		}
	}

	return nil
}
