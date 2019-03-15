package serviceaccount

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	AppFlags cmdapp.AppFlags
	Values   bool
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List service accounts",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	return cmd
}

func (o *ListOptions) Run() error {
	coreClient, err := o.depsFactory.CoreClient()
	if err != nil {
		return err
	}

	dynamicClient, err := o.depsFactory.DynamicClient()
	if err != nil {
		return err
	}

	app, err := ctlapp.NewApps(o.AppFlags.NamespaceFlags.Name, coreClient, dynamicClient).Find(o.AppFlags.Name)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	identifiedResources := ctlres.NewIdentifiedResources(coreClient, dynamicClient)

	rbacResources := &RBACResources{}

	resources, err := identifiedResources.List(labelSelector)
	if err != nil {
		return err
	}

	err = rbacResources.Collect(resources)
	if err != nil {
		return err
	}

	table := uitable.Table{
		Title:   "Service accounts",
		Content: "service accounts",

		Header: []uitable.Header{
			uitable.NewHeader("Namespace"),
			uitable.NewHeader("Name"),
			uitable.NewHeader("Binding/Role Namespace"),
			uitable.NewHeader("Binding"),
			uitable.NewHeader("Role"),
			// uitable.NewHeader("API Groups"),
			// uitable.NewHeader("Verbs"),
			// uitable.NewHeader("Resources"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true},
			{Column: 2, Asc: true},
			{Column: 3, Asc: true},
			{Column: 4, Asc: true},
		},
	}

	for _, roleBinding := range rbacResources.RoleBindings {
		var matchingSAs []*ServiceAccount
		var matchingRoles []*Role

		for _, sa := range rbacResources.ServiceAccounts {
			if roleBinding.MatchesServiceAccount(sa) {
				matchingSAs = append(matchingSAs, sa)
			}
		}

		for _, role := range rbacResources.Roles {
			if roleBinding.MatchesRole(role) {
				matchingRoles = append(matchingRoles, role)
				break
			}
		}

		for _, sa := range matchingSAs {
			for _, role := range matchingRoles {
				sa.MarkUsed()
				role.MarkUsed()
				roleBinding.MarkUsed()

				table.Rows = append(table.Rows, []uitable.Value{
					uitable.NewValueString(sa.Namespace()),
					uitable.NewValueString(sa.Name()),
					uitable.NewValueString(role.Namespace()),
					uitable.NewValueString(roleBinding.Name()),
					uitable.NewValueString(role.Name()),
					// uitable.NewValueStrings(role.APIGroups()),
					// cmdcore.NewValueStringsSingleLine(role.Verbs()),
					// uitable.NewValueStrings(role.Resources()),
				})
			}
		}
	}

	for _, sa := range rbacResources.ServiceAccounts {
		if !sa.Used() {
			table.Rows = append(table.Rows, []uitable.Value{
				uitable.NewValueString(sa.Namespace()),
				uitable.NewValueString(sa.Name()),
				uitable.NewValueString("?"),
				uitable.NewValueString("?"),
				uitable.NewValueString("?"),
			})
		}
	}

	for _, roleBinding := range rbacResources.RoleBindings {
		if !roleBinding.Used() {
			table.Rows = append(table.Rows, []uitable.Value{
				uitable.NewValueString("?"),
				uitable.NewValueString("?"),
				uitable.NewValueString(roleBinding.Namespace()),
				uitable.NewValueString(roleBinding.Name()),
				uitable.NewValueString("?"),
			})
		}
	}

	for _, role := range rbacResources.Roles {
		if !role.Used() {
			table.Rows = append(table.Rows, []uitable.Value{
				uitable.NewValueString("?"),
				uitable.NewValueString("?"),
				uitable.NewValueString(role.Namespace()),
				uitable.NewValueString("?"),
				uitable.NewValueString(role.Name()),
			})
		}
	}

	// TODO unused resources

	o.ui.PrintTable(table)

	return nil
}
