package tools

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type ListImagesOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags FileFlags
}

func NewListImagesOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ListImagesOptions {
	return &ListImagesOptions{ui: ui, depsFactory: depsFactory}
}

func NewListImagesCmd(o *ListImagesOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-images",
		Aliases: []string{"ls-images", "imgs"},
		Short:   "List images",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	return cmd
}

func (o *ListImagesOptions) Run() error {
	var resWithSources []ResourceWithSource

	for _, file := range o.FileFlags.Files {
		fileRs, err := ctlres.NewFileResources(file)
		if err != nil {
			return err
		}

		for _, fileRes := range fileRs {
			resources, err := fileRes.Resources()
			if err != nil {
				return err
			}

			for _, res := range resources {
				resWithSources = append(resWithSources, ResourceWithSource{res, fileRes.Description()})
			}
		}
	}

	sourceHeader := uitable.NewHeader("Source")
	sourceHeader.Hidden = true

	table := uitable.Table{
		Title:   "Images",
		Content: "images",

		Header: []uitable.Header{
			uitable.NewHeader("URL"),
			sourceHeader,
		},

		FillFirstColumn: true,
	}

	for _, resWithSrc := range resWithSources {
		for _, img := range findMapStringValues(resWithSrc.DeepCopyRaw(), "image") {
			table.Rows = append(table.Rows, []uitable.Value{
				uitable.NewValueString(img),
				uitable.NewValueString(resWithSrc.Source),
			})
		}
	}

	o.ui.PrintTable(table)

	return nil
}

type ResourceWithSource struct {
	ctlres.Resource
	Source string
}

func findMapStringValues(obj interface{}, key string) []string {
	var result []string

	switch typedObj := obj.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			if k == key {
				if typedValue, ok := v.(string); ok {
					result = append(result, typedValue)
					continue
				}
			}
			result = append(result, findMapStringValues(typedObj[k], key)...)
		}
	case map[string]string:
		for k, v := range typedObj {
			if k == key {
				result = append(result, v)
			}
		}
	case []interface{}:
		for _, o := range typedObj {
			result = append(result, findMapStringValues(o, key)...)
		}
	}

	return result
}
