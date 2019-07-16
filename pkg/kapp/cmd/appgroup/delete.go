package appgroup

import (
	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	cmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	"github.com/spf13/cobra"
)

const (
	appGroupAnnKey = "kapp.k14s.io/app-group"
)

type DeleteOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppGroupFlags AppGroupFlags
	DeployFlags   DeployFlags
	AppFlags      DeleteAppFlags
}

type DeleteAppFlags struct {
	DiffFlags  cmdtools.DiffFlags
	ApplyFlags cmdapp.ApplyFlags
}

func NewDeleteOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *DeleteOptions {
	return &DeleteOptions{ui: ui, depsFactory: depsFactory}
}

func NewDeleteCmd(o *DeleteOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deploy app group",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppGroupFlags.Set(cmd, flagsFactory)
	o.AppFlags.DiffFlags.SetWithPrefix("diff", cmd)
	o.AppFlags.ApplyFlags.SetWithDefaults("", cmdapp.ApplyFlagsDeleteDefaults, cmd)
	return cmd
}

func (o *DeleteOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	dynamicClient, err := o.depsFactory.DynamicClient()
	if err != nil {
		return err
	}

	apps := ctlapp.NewApps(o.AppGroupFlags.NamespaceFlags.Name, coreClient, dynamicClient)

	appsInGroup, err := apps.List(map[string]string{appGroupAnnKey: o.AppGroupFlags.Name})
	if err != nil {
		return err
	}

	for _, app := range appsInGroup {
		err := o.deleteApp(app.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *DeleteOptions) deleteApp(name string) error {
	o.ui.PrintLinef("--- deleting app '%s' (namespace: %s)",
		name, o.AppGroupFlags.NamespaceFlags.Name)

	deleteOpts := cmdapp.NewDeleteOptions(o.ui, o.depsFactory)
	deleteOpts.AppFlags = cmdapp.AppFlags{
		Name:           name,
		NamespaceFlags: o.AppGroupFlags.NamespaceFlags,
	}
	deleteOpts.DiffFlags = o.AppFlags.DiffFlags
	deleteOpts.ApplyFlags = o.AppFlags.ApplyFlags

	return deleteOpts.Run()
}
