// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"github.com/spf13/cobra"
)

type FileFlags struct {
	Files []string
	Sort  bool
}

func (s *FileFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.Files, "file", "f", s.Files, "Set file (format: /tmp/foo, https://..., -) (can repeat)")
	cmd.Flags().BoolVar(&s.Sort, "sort", true, "Sort by namespace, name, etc.")
}

type FileFlags2 struct {
	Files []string
}

func (s *FileFlags2) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.Files, "file2", nil, "Set second file (format: /tmp/foo, https://..., -) (can repeat)")
}
