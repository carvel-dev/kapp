// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	cmdcore "carvel.dev/kapp/pkg/kapp/cmd/core"
	"github.com/spf13/cobra"
)

type ProfilingFlags struct {
	profileName   string
	profileOutput string
}

func (p *ProfilingFlags) Set(cmd *cobra.Command, _ cmdcore.FlagsFactory) {
	cmd.PersistentFlags().StringVar(&p.profileName, "profile", "none", "Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex)")
	cmd.PersistentFlags().StringVar(&p.profileOutput, "profile-output", "profile.pprof", "Name of the file to write the profile to")
}

func (p *ProfilingFlags) initProfiling() error {
	var (
		f   *os.File
		err error
	)
	switch p.profileName {
	case "none":
		return nil
	case "cpu":
		f, err = os.Create(p.profileOutput)
		if err != nil {
			return err
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return err
		}
	// Block and mutex profiles need a call to Set{Block,Mutex}ProfileRate to
	// output anything. We choose to sample all events.
	case "block":
		runtime.SetBlockProfileRate(1)
	case "mutex":
		runtime.SetMutexProfileFraction(1)
	default:
		// Check the profile name is valid.
		if profile := pprof.Lookup(p.profileName); profile == nil {
			return fmt.Errorf("unknown profile '%s'", p.profileName)
		}
	}

	// If the command is interrupted before the end (ctrl-c), flush the
	// profiling files
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		f.Close()
		p.flushProfiling()
		os.Exit(0)
	}()

	return nil
}

func (p *ProfilingFlags) flushProfiling() error {
	switch p.profileName {
	case "none":
		return nil
	case "cpu":
		pprof.StopCPUProfile()
	case "heap":
		runtime.GC()
		fallthrough
	default:
		profile := pprof.Lookup(p.profileName)
		if profile == nil {
			return nil
		}
		f, err := os.Create(p.profileOutput)
		if err != nil {
			return err
		}
		defer f.Close()
		return profile.WriteTo(f, 0)
	}

	return nil
}
