// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// Taken from https://github.com/linkerd/linkerd2/blob/main/cli/cmd/completion.go

func NewCmdCompletion() *cobra.Command {
	example := `  # bash <= 3.2
  source /dev/stdin <<< "$(kapp completion bash)"

  # bash >= 4.0
  source <(kapp completion bash)

  # bash <= 3.2 on osx
  brew install bash-completion # ensure you have bash-completion 1.3+
  kapp completion bash > $(brew --prefix)/etc/bash_completion.d/kapp

  # bash >= 4.0 on osx
  brew install bash-completion@2
  kapp completion bash > $(brew --prefix)/etc/bash_completion.d/kapp

  # zsh
  source <(kapp completion zsh)

  # zsh on osx / oh-my-zsh
  kapp completion zsh > "${fpath[1]}/_kapp"

  # fish:
  kapp completion fish | source

  # To load completions for each session, execute once:
  kapp completion fish > ~/.config/fish/completions/kapp.fish

  # powershell:
  kapp completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  kapp completion powershell > kapp.ps1
  # and source this file from your powershell profile.
`

	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Output shell completion code for the specified shell (bash, zsh or fish)",
		Long:      "Output shell completion code for the specified shell (bash, zsh or fish).",
		Example:   example,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := getCompletion(args[0], cmd.Parent())
			if err != nil {
				return err
			}

			fmt.Print(out)
			return nil
		},
	}

	return cmd
}

// getCompletion will return the auto completion shell script, if supported
func getCompletion(sh string, parent *cobra.Command) (string, error) {
	var err error
	var buf bytes.Buffer

	switch sh {
	case "bash":
		err = parent.GenBashCompletion(&buf)
	case "zsh":
		err = parent.GenZshCompletion(&buf)
	case "fish":
		err = parent.GenFishCompletion(&buf, true)
	case "powershell":
		err = parent.GenPowerShellCompletion(&buf)
	default:
		err = errors.New("unsupported shell type (must be bash, zsh or fish): " + sh)
	}

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
