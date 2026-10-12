// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetCompletion(t *testing.T) {
	exs := []getCompletionExample{
		{Shell: "bash"},
		{Shell: "zsh"},
		{Shell: "fish"},
		{Shell: "powershell"},
		{
			Shell:       "tcsh",
			ExpectedErr: "Unsupported shell type (must be bash, zsh, fish or powershell): tcsh",
		},
	}

	for _, ex := range exs {
		ex.Check(t)
	}
}

type getCompletionExample struct {
	Shell       string
	ExpectedErr string
}

func (e getCompletionExample) Check(t *testing.T) {
	out, err := getCompletion(e.Shell, &cobra.Command{Use: "kapp"})
	if len(e.ExpectedErr) > 0 {
		require.EqualError(t, err, e.ExpectedErr, "shell=%s", e.Shell)
		return
	}

	require.NoError(t, err, "shell=%s", e.Shell)
	require.NotEmpty(t, out, "shell=%s", e.Shell)
}
