package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kapp/pkg/kapp/cmd"

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
		confUI.ErrorLinef("Error: %v", err)
		os.Exit(1)
	}

	confUI.PrintLinef("Succeeded")
}
