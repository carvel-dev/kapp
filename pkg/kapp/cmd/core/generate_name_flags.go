// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GenerateNameFlags struct {
	GenerateName bool
}

func (s *GenerateNameFlags) Set(cmd *cobra.Command, _ FlagsFactory) {
	cmd.Flags().BoolVar(&s.GenerateName, "generate-name", false, "Set to generate name")
}

func (s *GenerateNameFlags) Apply(meta metav1.ObjectMeta) metav1.ObjectMeta {
	if s.GenerateName {
		meta.GenerateName = meta.Name + "-"
		meta.Name = ""
	}
	return meta
}
