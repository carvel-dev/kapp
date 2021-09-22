// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlcap "github.com/k14s/kapp/pkg/kapp/clusterapply"
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
	newResources, err := o.fileResources(o.FileFlags.Files)
	if err != nil {
		return err
	}

	existingResources, err := o.fileResources(o.FileFlags2.Files)
	if err != nil {
		return err
	}

	changeFactory := ctldiff.NewChangeFactory(nil, nil)
	changes, err := ctldiff.NewChangeSet(existingResources, newResources, o.DiffFlags.ChangeSetOpts, changeFactory).Calculate()
	if err != nil {
		return err
	}
	var changeViews []ctlcap.ChangeView

	for _, change := range changes {
		changeViews = append(changeViews, DiffChangeView{change})
	}

	// TODO support adding custom config for mask rules?
	ctlcap.NewChangeSetView(changeViews, nil, o.DiffFlags.ChangeSetViewOpts).Print(o.ui)

	return nil
}

func (o *DiffOptions) fileResources(files []string) ([]ctlres.Resource, error) {
	var newResources []ctlres.Resource

	for _, file := range files {
		fileRs, err := ctlres.NewFileResources(file)
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

type DiffChangeView struct {
	change ctldiff.Change
}

var _ ctlcap.ChangeView = DiffChangeView{}

func (v DiffChangeView) Resource() ctlres.Resource         { return v.change.NewOrExistingResource() }
func (v DiffChangeView) ExistingResource() ctlres.Resource { return v.change.ExistingResource() }

func (v DiffChangeView) ApplyOp() ctlcap.ClusterChangeApplyOp {
	switch v.change.Op() {
	case ctldiff.ChangeOpAdd:
		return ctlcap.ClusterChangeApplyOpAdd
	case ctldiff.ChangeOpDelete:
		return ctlcap.ClusterChangeApplyOpDelete
	case ctldiff.ChangeOpUpdate:
		return ctlcap.ClusterChangeApplyOpUpdate
	case ctldiff.ChangeOpKeep:
		return ctlcap.ClusterChangeApplyOpNoop
	default:
		panic("Unknown change apply op")
	}
}

func (v DiffChangeView) ApplyStrategyOp() (ctlcap.ClusterChangeApplyStrategyOp, error) {
	return ctlcap.UnknownStrategyOp, nil
}

// Since we are diffing changes without a cluster, there will be no wait operations
func (v DiffChangeView) WaitOp() ctlcap.ClusterChangeWaitOp { return ctlcap.ClusterChangeWaitOpNoop }

func (v DiffChangeView) ConfigurableTextDiff() *ctldiff.ConfigurableTextDiff {
	return v.change.ConfigurableTextDiff()
}
