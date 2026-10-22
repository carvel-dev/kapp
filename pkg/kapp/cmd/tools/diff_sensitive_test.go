// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
)

type diffCaptureUI struct {
	ui.UI
	output bytes.Buffer
}

func (u *diffCaptureUI) PrintBlock(b []byte)                      { u.output.Write(b) }
func (u *diffCaptureUI) BeginLinef(f string, args ...interface{}) { fmt.Fprintf(&u.output, f, args...) }
func TestToolsDiffMasksSecrets(t *testing.T) {
	for _, yamlOutput := range []bool{false, true} {
		for _, mask := range []bool{false, true} {
			capture := &diffCaptureUI{UI: ui.NewNoopUI()}
			opts := NewDiffOptions(capture, nil)
			opts.DiffFlags.SetWithPrefix("", &cobra.Command{})
			opts.DiffFlags.Changes = !yamlOutput
			opts.DiffFlags.ChangesYAML = yamlOutput
			opts.DiffFlags.Mask = mask
			opts.DiffFlags.Summary = false
			content := func(v string) []byte {
				return []byte("apiVersion: v1\nkind: Secret\nmetadata:\n  name: credentials\ndata:\n  password: " + base64.StdEncoding.EncodeToString([]byte(v)) + "\nstringData:\n  token: " + v + "\n")
			}
			opts.FileSystem = fstest.MapFS{"old.yml": &fstest.MapFile{Data: content("old-canary")}, "new.yml": &fstest.MapFile{Data: content("new-canary")}}
			opts.FileFlags.Files = []string{"new.yml"}
			opts.FileFlags2.Files = []string{"old.yml"}
			if err := opts.Run(); err != nil {
				t.Fatal(err)
			}
			for _, marker := range []string{"new-canary", base64.StdEncoding.EncodeToString([]byte("new-canary"))} {
				found := strings.Contains(capture.output.String(), marker)
				if found == mask {
					t.Fatalf("mask=%v yaml=%v: unexpected credential visibility", mask, yamlOutput)
				}
			}
		}
	}
}
