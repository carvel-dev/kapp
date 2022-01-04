// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DescribeOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	appChange string

	AppFlags       cmdapp.Flags
	NamespaceFlags cmdcore.NamespaceFlags
}

func NewDescribeOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *DescribeOptions {
	return &DescribeOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewDescribeCmd(o *DescribeOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe",
		Aliases: []string{"d"},
		Short:   "View changes applied for app-change",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	cmd.Flags().StringVar(&o.appChange, "app-change", "", "Name of app-change to be described")
	return cmd
}

func (o *DescribeOptions) Run() error {
	if o.appChange == "" {
		return fmt.Errorf("Flag --app-change is required")
	}

	client, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	appChangeMap, err := client.CoreV1().ConfigMaps(o.AppFlags.NamespaceFlags.Name).Get(context.TODO(), o.appChange, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, found := appChangeMap.Labels[ctlapp.IsChangeLabelKey]
	if !found {
		return fmt.Errorf("Not an app-change")
	}

	var configSpec ctlapp.ChangeMeta
	err = json.Unmarshal([]byte(appChangeMap.Data[ctlapp.ChangeMetaKey]), &configSpec)
	if err != nil {
		return fmt.Errorf("Unmarshalling app-change spec: %s", err.Error())
	}

	if configSpec.DiffChanges == "" {
		o.ui.PrintLinef("Diff changes have not been stored for this app-change")
		return nil
	}

	o.ui.PrintLinef(configSpec.DiffChanges)

	return nil
}
