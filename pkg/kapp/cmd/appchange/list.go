// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package appchange

import (
	"fmt"
	"strings"
	"time"

	ctlapp "carvel.dev/kapp/pkg/kapp/app"
	cmdapp "carvel.dev/kapp/pkg/kapp/cmd/app"
	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"carvel.dev/kapp/pkg/kapp/logger"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory
	logger      logger.Logger

	AppFlags  cmdapp.Flags
	TimeFlags TimeFlags
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory, logger logger.Logger) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory, logger: logger}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List app changes",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.AppFlags.Set(cmd, flagsFactory)
	o.TimeFlags.Set(cmd)
	return cmd
}

func (o *ListOptions) Run() error {
	app, _, err := cmdapp.Factory(o.depsFactory, o.AppFlags, cmdapp.ResourceTypesFlags{}, o.logger)
	if err != nil {
		return err
	}

	changes, err := app.Changes()
	if err != nil {
		return err
	}

	formats := []string{time.RFC3339, "2006-01-02"}

	if o.TimeFlags.Before != "" {
		o.TimeFlags.BeforeTime, err = o.parseTime(o.TimeFlags.Before, formats)
		if err != nil {
			return err
		}
	}

	if o.TimeFlags.After != "" {
		o.TimeFlags.AfterTime, err = o.parseTime(o.TimeFlags.After, formats)
		if err != nil {
			return err
		}

		if !o.TimeFlags.BeforeTime.IsZero() && o.TimeFlags.BeforeTime.Before(o.TimeFlags.AfterTime) {
			return fmt.Errorf("After time %s should be less than before time %s", o.TimeFlags.After, o.TimeFlags.Before)
		}
	}

	AppChangesTable{"App changes", changes, o.TimeFlags}.Print(o.ui)

	return nil
}

func (o *ListOptions) parseTime(input string, formats []string) (time.Time, error) {
	for _, format := range formats {
		t, err := time.Parse(format, input)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format %s, supported formats: %s", input, formats)
}

type AppChangesTable struct {
	Title     string
	Changes   []ctlapp.Change
	TimeFlags TimeFlags
}

func (t AppChangesTable) Print(ui ui.UI) {
	nsHeader := uitable.NewHeader("Namespaces")
	nsHeader.Hidden = true

	table := uitable.Table{
		Title:   t.Title,
		Content: "app changes",

		Header: []uitable.Header{
			uitable.NewHeader("Name"),
			uitable.NewHeader("Started At"),
			uitable.NewHeader("Finished At"),
			uitable.NewHeader("Successful"),
			uitable.NewHeader("Description"),
			nsHeader,
		},

		SortBy: []uitable.ColumnSort{
			{Column: 1, Asc: false},
			{Column: 0, Asc: true}, // in case start time are same
		},
	}

	for _, change := range t.Changes {
		if (!t.TimeFlags.BeforeTime.IsZero() && !change.Meta().StartedAt.Before(t.TimeFlags.BeforeTime)) ||
			!change.Meta().StartedAt.After(t.TimeFlags.AfterTime) {
			continue
		}

		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(change.Name()),
			uitable.NewValueTime(change.Meta().StartedAt),
			uitable.NewValueTime(change.Meta().FinishedAt),
			uitable.ValueFmt{
				V:     cmdcore.NewValueUnknownBool(change.Meta().Successful),
				Error: change.Meta().Successful == nil || *change.Meta().Successful != true,
			},
			uitable.NewValueString(change.Meta().Description),
			uitable.NewValueString(strings.Join(change.Meta().Namespaces, ",")),
		})
	}

	ui.PrintTable(table)
}
