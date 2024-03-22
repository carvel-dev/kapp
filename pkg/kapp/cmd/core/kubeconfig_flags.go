// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"os"

	"github.com/cppforlife/cobrautil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type KubeconfigFlags struct {
	Path    *KubeconfigPathFlag
	Context *KubeconfigContextFlag
	YAML    *KubeconfigYAMLFlag
}

func (f *KubeconfigFlags) Set(cmd *cobra.Command, _ FlagsFactory) {
	f.Path = NewKubeconfigPathFlag()
	cmd.PersistentFlags().Var(f.Path, "kubeconfig", "Path to the kubeconfig file ($KAPP_KUBECONFIG)")

	f.Context = NewKubeconfigContextFlag()
	cmd.PersistentFlags().Var(f.Context, "kubeconfig-context", "Kubeconfig context override ($KAPP_KUBECONFIG_CONTEXT)")

	f.YAML = NewKubeconfigYAMLFlag()
	cmd.PersistentFlags().Var(f.YAML, "kubeconfig-yaml", "Kubeconfig contents as YAML ($KAPP_KUBECONFIG_YAML)")
}

type KubeconfigPathFlag struct {
	value string
}

var _ pflag.Value = &KubeconfigPathFlag{}
var _ cobrautil.ResolvableFlag = &KubeconfigPathFlag{}

func NewKubeconfigPathFlag() *KubeconfigPathFlag {
	return &KubeconfigPathFlag{}
}

func (s *KubeconfigPathFlag) Set(val string) error {
	s.value = val
	return nil
}

func (s *KubeconfigPathFlag) Type() string   { return "string" }
func (s *KubeconfigPathFlag) String() string { return "" } // default for usage

func (s *KubeconfigPathFlag) Value() (string, error) {
	err := s.Resolve()
	if err != nil {
		return "", err
	}

	return s.value, nil
}

func (s *KubeconfigPathFlag) Resolve() error {
	if len(s.value) > 0 {
		return nil
	}

	s.value = s.resolveValue()

	return nil
}

func (s *KubeconfigPathFlag) resolveValue() string {
	path := os.Getenv("KAPP_KUBECONFIG")
	if len(path) > 0 {
		return path
	}

	return ""
}

type KubeconfigContextFlag struct {
	value string
}

var _ pflag.Value = &KubeconfigContextFlag{}
var _ cobrautil.ResolvableFlag = &KubeconfigPathFlag{}

func NewKubeconfigContextFlag() *KubeconfigContextFlag {
	return &KubeconfigContextFlag{}
}

func (s *KubeconfigContextFlag) Set(val string) error {
	s.value = val
	return nil
}

func (s *KubeconfigContextFlag) Type() string   { return "string" }
func (s *KubeconfigContextFlag) String() string { return "" } // default for usage

func (s *KubeconfigContextFlag) Value() (string, error) {
	err := s.Resolve()
	if err != nil {
		return "", err
	}

	return s.value, nil
}

func (s *KubeconfigContextFlag) Resolve() error {
	if len(s.value) > 0 {
		return nil
	}

	s.value = os.Getenv("KAPP_KUBECONFIG_CONTEXT")

	return nil
}

type KubeconfigYAMLFlag struct {
	value string
}

var _ pflag.Value = &KubeconfigYAMLFlag{}
var _ cobrautil.ResolvableFlag = &KubeconfigPathFlag{}

func NewKubeconfigYAMLFlag() *KubeconfigYAMLFlag {
	return &KubeconfigYAMLFlag{}
}

func (s *KubeconfigYAMLFlag) Set(val string) error {
	s.value = val
	return nil
}

func (s *KubeconfigYAMLFlag) Type() string   { return "string" }
func (s *KubeconfigYAMLFlag) String() string { return "" } // default for usage

func (s *KubeconfigYAMLFlag) Value() (string, error) {
	err := s.Resolve()
	if err != nil {
		return "", err
	}

	return s.value, nil
}

func (s *KubeconfigYAMLFlag) Resolve() error {
	if len(s.value) > 0 {
		return nil
	}

	s.value = os.Getenv("KAPP_KUBECONFIG_YAML")

	return nil
}
