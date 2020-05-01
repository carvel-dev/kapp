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
	rand.Seed(time.Now().UTC().UnixNano())

	// TODO logs
	// TODO log flags used

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	command := cmd.NewDefaultKappCmd(confUI)

	err := command.Execute()
	if err != nil {
		confUI.ErrorLinef("kapp: Error: %v", uierrs.NewMultiLineError(err))
		if typedErr, ok := err.(cmdapp.ExitStatus); ok {
			os.Exit(typedErr.ExitStatus())
		}
		os.Exit(1)
	}

	confUI.PrintLinef("Succeeded")
}
