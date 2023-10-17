// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttresmod

import (
	"fmt"

	"github.com/ghodss/yaml"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

type OverlayContractV1Mod struct {
	ResourceMatcher ctlres.ResourceMatcher
	OverlayYAML     string

	// TODO support rebase_resource(res, sources) func via .star file?
	// Starlark string
}

var _ ctlres.ResourceModWithMultiple = OverlayContractV1Mod{}

func (t OverlayContractV1Mod) IsResourceMatching(res ctlres.Resource) bool {
	if res == nil || !t.ResourceMatcher.Matches(res) {
		return false
	}
	return true
}

func (t OverlayContractV1Mod) ApplyFromMultiple(res ctlres.Resource, srcs map[ctlres.FieldCopyModSource]ctlres.Resource) error {
	result, err := t.evalYtt(res, srcs)
	if err != nil {
		return fmt.Errorf("Applying ytt (overlayContractV1): %w", err)
	}

	res.DeepCopyIntoFrom(result)
	return nil
}

func (t OverlayContractV1Mod) evalYtt(res ctlres.Resource, srcs map[ctlres.FieldCopyModSource]ctlres.Resource) (ctlres.Resource, error) {
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
		return nil, err
	}

	filesToProcess := []*files.File{
		// Current resource we are working with:
		files.MustNewFileFromSource(files.NewBytesSource("resource.yml", resYAMLBs)),
		// Overlay instructions
		files.MustNewFileFromSource(files.NewBytesSource("overlay.yml", []byte(t.OverlayYAML))),
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	if out.Err != nil {
		return nil, fmt.Errorf("Evaluating: %w", out.Err)
	}

	if len(out.Files) == 0 {
		return nil, fmt.Errorf("Expected to find resource.yml but saw zero files")
	}

	file := out.Files[0]
	if file.RelativePath() != "resource.yml" {
		return nil, fmt.Errorf("Expected resource.yml but was: %s", file.RelativePath())
	}

	rs, err := ctlres.NewResourcesFromBytes(file.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Deserializing result: %w", err)
	}

	if len(rs) != 1 {
		return nil, fmt.Errorf("Expected one resource to be returned from ytt, but was %d", len(rs))
	}

	return rs[0], nil
}

func (t OverlayContractV1Mod) valuesYAML(srcs map[ctlres.FieldCopyModSource]ctlres.Resource) ([]byte, error) {
	values := map[string]interface{}{}
	for src, res := range srcs {
		if res != nil {
			values[string(src)] = res.DeepCopyRaw()
		} else {
			values[string(src)] = nil
		}
	}
	return yaml.Marshal(values)
}
