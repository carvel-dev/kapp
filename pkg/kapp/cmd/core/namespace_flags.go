package core

import (
	"fmt"
	"os"

	"github.com/cppforlife/cobrautil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type NamespaceFlags struct {
	Name string
}

func (s *NamespaceFlags) Set(cmd *cobra.Command, flagsFactory FlagsFactory) {
	name := flagsFactory.NewNamespaceNameFlag(&s.Name)
	cmd.Flags().VarP(name, "namespace", "n", "Specified namespace ($KAPP_NAMESPACE or default from kubeconfig)")
}

type NamespaceNameFlag struct {
	value         *string
	configFactory ConfigFactory
}

var _ pflag.Value = &NamespaceNameFlag{}
var _ cobrautil.ResolvableFlag = &NamespaceNameFlag{}

func NewNamespaceNameFlag(value *string, configFactory ConfigFactory) *NamespaceNameFlag {
	return &NamespaceNameFlag{value, configFactory}
}

func (s *NamespaceNameFlag) Set(val string) error {
	*s.value = val
	return nil
}

func (s *NamespaceNameFlag) Type() string   { return "string" }
func (s *NamespaceNameFlag) String() string { return "" } // default for usage

func (s *NamespaceNameFlag) Resolve() error {
	value, err := s.resolveValue()
	if err != nil {
		return err
	}

	*s.value = value

	return nil
}

func (s *NamespaceNameFlag) resolveValue() (string, error) {
	if s.value != nil && len(*s.value) > 0 {
		return *s.value, nil
	}

	envVal := os.Getenv("KAPP_NAMESPACE")
	if len(envVal) > 0 {
		return envVal, nil
	}

	configVal, err := s.configFactory.DefaultNamespace()
	if err != nil {
		return configVal, nil
	}

	if len(configVal) > 0 {
		return configVal, nil
	}

	return "", fmt.Errorf("Expected to non-empty namespace name")
}
