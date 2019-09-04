package appchange

import (
	"github.com/cppforlife/go-cli-ui/ui"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type GCOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags cmdapp.AppFlags
	Max      int
}

func NewGCOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *GCOptions {
	return &GCOptions{ui: ui, depsFactory: depsFactory}
}

func NewGCCmd(o *GCOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "garbare-collect",
		Aliases: []string{"gc"},
		Short:   "Garbage collect app changes",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	cmd.Flags().IntVar(&o.Max, "max", 200, "Maximum number of app changes to keep")
	return cmd
}

func (o *GCOptions) Run() error {
	app, _, _, err := cmdapp.AppFactory(o.depsFactory, o.AppFlags)
	if err != nil {
		return err
	}

	changes, err := app.Changes()
	if err != nil {
		return err
	}

	if len(changes) < o.Max {
		o.ui.PrintLinef("Keeping %d app changes", len(changes))
		return nil
	}

	changes = changes[0 : len(changes)-o.Max]

	AppChangesTable{"App changes to delete", changes}.Print(o.ui)

	err = o.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	for _, change := range changes {
		err := change.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}
