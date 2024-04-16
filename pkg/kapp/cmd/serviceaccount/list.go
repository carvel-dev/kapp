// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package serviceaccount

import (
	cmdapp "carvel.dev/kapp/pkg/kapp/cmd/app"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags cmdapp.Flags
	Values   bool
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory, logger: logger}
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
	app, supportObjs, err := cmdapp.Factory(o.depsFactory, o.AppFlags, cmdapp.ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	labelSelector, err := app.LabelSelector()
	if err != nil {
		return err
	}

	rbacResources := &RBACResources{}

	resources, err := supportObjs.IdentifiedResources.List(labelSelector, nil, resources.IdentifiedResourcesListOpts{})
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
					cmdcore.NewValueNamespace(sa.Namespace()),
					uitable.NewValueString(sa.Name()),
					cmdcore.NewValueNamespace(role.Namespace()),
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
				cmdcore.NewValueNamespace(sa.Namespace()),
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
				cmdcore.NewValueNamespace(roleBinding.Namespace()),
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
				cmdcore.NewValueNamespace(role.Namespace()),
				uitable.NewValueString("?"),
				uitable.NewValueString(role.Name()),
			})
		}
	}

	// TODO unused resources

	o.ui.PrintTable(table)

	return nil
}
