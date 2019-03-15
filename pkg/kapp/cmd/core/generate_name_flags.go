package core

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GenerateNameFlags struct {
	GenerateName bool
}

func (s *GenerateNameFlags) Set(cmd *cobra.Command, flagsFactory FlagsFactory) {
	cmd.Flags().BoolVar(&s.GenerateName, "generate-name", false, "Set to generate name")
}

func (s *GenerateNameFlags) Apply(meta metav1.ObjectMeta) metav1.ObjectMeta {
	if s.GenerateName {
		meta.GenerateName = meta.Name + "-"
		meta.Name = ""
	}
	return meta
}
