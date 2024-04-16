// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"sigs.k8s.io/yaml"
)

type WaitRuleContractV1 struct {
	ResourceMatcher ctlres.ResourceMatcher
	Starlark        string
}

type waitRuleContractV1Result struct {
	Result WaitRuleContractV1ResultDetails
}

type WaitRuleContractV1ResultDetails struct {
	Done       bool   `json:"done"`
	Successful bool   `json:"successful"`
	Message    string `json:"message"`

	UnblockChanges bool `json:"unblockChanges"`
}

func (t WaitRuleContractV1) Apply(res ctlres.Resource) (*WaitRuleContractV1ResultDetails, error) {
	if !t.ResourceMatcher.Matches(res) {
		return nil, nil
	}

	return t.evalYtt(res)
}

func (t WaitRuleContractV1) evalYtt(res ctlres.Resource) (*WaitRuleContractV1ResultDetails, error) {
	opts := cmdtpl.NewOptions()

	opts.DataValuesFlags.FromFiles = []string{"values.yml"}
	opts.DataValuesFlags.ReadFileFunc = func(path string) ([]byte, error) {
		if path != "values.yml" {
			return nil, fmt.Errorf("Unknown file to read: %s", path)
		}
		return yaml.Marshal(res.DeepCopyRaw())
	}

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("resource.star", []byte(t.Starlark))),
		files.MustNewFileFromSource(files.NewBytesSource("config.yml", t.getConfigYAML())),
	}

	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	if out.Err != nil {
		return nil, fmt.Errorf("Evaluating: %w", out.Err)
	}

	if len(out.Files) == 0 {
		fmt.Printf("Expected to find config.yml but saw zero files")
	}

	file := out.Files[0]
	if file.RelativePath() != "config.yml" {
		fmt.Printf("Expected config.yml but was: %s", file.RelativePath())
	}

	configObj := waitRuleContractV1Result{}

	err := yaml.Unmarshal(file.Bytes(), &configObj)
	if err != nil {
		return nil, fmt.Errorf("Deserializing result: %w", err)
	}

	return &configObj.Result, nil
}

func (t WaitRuleContractV1) getConfigYAML() []byte {
	config := `
#@ load("resource.star", "is_done")
#@ load("@ytt:data", "data")

result: #@ is_done(data.values)
`
	return []byte(config)
}
