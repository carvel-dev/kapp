package yttresmod

import (
	"fmt"

	"github.com/ghodss/yaml"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
)

type Mod struct {
	ResourceMatcher ctlres.ResourceMatcher
	TemplateYAML    string

	// TODO support rebase_resource(res, sources) func via .star file?
	// Starlark string
}

var _ ctlres.ResourceModWithMultiple = Mod{}

func (t Mod) ApplyFromMultiple(res ctlres.Resource, srcs map[ctlres.FieldCopyModSource]ctlres.Resource) error {
	if !t.ResourceMatcher.Matches(res) {
		return nil
	}

	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags.FromFiles = []string{"values.yml"}
	opts.DataValuesFlags.ReadFileFunc = func(path string) ([]byte, error) {
		if path != "values.yml" {
			return nil, fmt.Errorf("Unknown file to read: %s", path)
		}
		return t.valuesYAML(srcs)
	}

	resYAMLBs, err := res.AsYAMLBytes()
	if err != nil {
		return err
	}

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("resource.yml", resYAMLBs)),
		files.MustNewFileFromSource(files.NewBytesSource("template.yml", []byte(t.TemplateYAML))),
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	if out.Err != nil {
		return out.Err
	}

	if len(out.Files) == 0 {
		return fmt.Errorf("Expected to find resource.yml but saw zero files")
	}

	file := out.Files[0]
	if file.RelativePath() != "resource.yml" {
		return fmt.Errorf("Expected resource.yml but was: %s", file.RelativePath())
	}

	rs, err := ctlres.NewResourcesFromBytes(file.Bytes())
	if err != nil {
		return err
	}

	if len(rs) != 1 {
		return fmt.Errorf("Expected exactly one resource to be returned from ytt templates")
	}

	res.DeepCopyIntoFrom(rs[0])
	return nil
}

func (t Mod) valuesYAML(srcs map[ctlres.FieldCopyModSource]ctlres.Resource) ([]byte, error) {
	values := map[string]interface{}{}
	for src, res := range srcs {
		values[string(src)] = res.DeepCopyRaw()
	}
	return yaml.Marshal(values)
}
