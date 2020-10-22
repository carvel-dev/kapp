// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"math/rand"
	"os"
	"time"

	uierrs "github.com/cppforlife/go-cli-ui/errors"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kapp/pkg/kapp/cmd"
	cmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	err := nonExitingMain()
	if err != nil {
		if typedErr, ok := err.(cmdapp.ExitStatus); ok {
			os.Exit(typedErr.ExitStatus())
		}
		os.Exit(1)
	}
}

// nonExitingMain does not use os.Exit to make sure Go runs defers
func nonExitingMain() error {
	rand.Seed(time.Now().UTC().UnixNano())

	// TODO logs
	// TODO log flags used

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	command := cmd.NewDefaultKappCmd(confUI)

	err := command.Execute()
	if err != nil {
		confUI.ErrorLinef("kapp: Error: %v", uierrs.NewMultiLineError(err))
		return err
	}

	confUI.PrintLinef("Succeeded")
	return nil
}
