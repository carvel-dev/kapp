// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttresmod

import (
	"fmt"

	"github.com/ghodss/yaml"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
)

type WaitRuleContractV1 struct {
	ResourceMatcher ctlres.ResourceMatcher
	Starlark        string
}

type ConfigYAMLObj struct {
	Result DoneApplyState
}

type DoneApplyState struct {
	Done       bool
	Successful bool
	Message    string
}

func (t WaitRuleContractV1) ApplyYttWaitRule(res ctlres.Resource) (*DoneApplyState, error) {
	if !t.ResourceMatcher.Matches(res) {
		return nil, nil
	}

	return t.evalYtt(res)
}

func (t WaitRuleContractV1) evalYtt(res ctlres.Resource) (*DoneApplyState, error) {
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags.FromFiles = []string{"values.yml"}
	opts.DataValuesFlags.ReadFileFunc = func(path string) ([]byte, error) {
		if path != "values.yml" {
			return nil, fmt.Errorf("Unknown file to read: %s", path)
		}
		return t.valuesYAML(res)
	}

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("rules.star", []byte(t.Starlark))),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", t.getConfigYAML())),
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	if out.Err != nil {
		return nil, fmt.Errorf("Evaluating: %s", out.Err)
	}

	if len(out.Files) == 0 {
		fmt.Printf("Expected to find config.yml but saw zero files")
	}

	file := out.Files[0]
	if file.RelativePath() != "config.yml" {
		fmt.Printf("Expected config.yml but was: %s", file.RelativePath())
	}

	configObj := ConfigYAMLObj{}

	err := yaml.Unmarshal(file.Bytes(), &configObj)
	if err != nil {
		return nil, fmt.Errorf("Deserializing result: %s", err)
	}

	return &configObj.Result, nil
}

func (t WaitRuleContractV1) valuesYAML(res ctlres.Resource) ([]byte, error) {
	return yaml.Marshal(res.DeepCopyRaw())
}

func (t WaitRuleContractV1) getConfigYAML() []byte {
	config := `
#@ load("rules.star", "check_status")
#@ load("@ytt:data", "data")

result: #@ check_status(data.values)
`
	return []byte(config)
}
