// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceTypesFlags struct {
	IgnoreFailingAPIServices   bool
	CanIgnoreFailingAPIService func(schema.GroupVersion) bool

	ScopeToFallbackAllowedNamespaces bool
}

func (s *ResourceTypesFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.IgnoreFailingAPIServices, "dangerous-ignore-failing-api-services",
		false, "Allow to ignore failing APIServices")

	cmd.Flags().BoolVar(&s.ScopeToFallbackAllowedNamespaces, "dangerous-scope-to-fallback-allowed-namespaces",
		true, "Scope resource searching to fallback allowed namespaces")

	cmd.Flags().MarkDeprecated("dangerous-scope-to-fallback-allowed-namespaces", "The default and recommended behavior is to use this fallback mechanism now")
}

func (s *ResourceTypesFlags) FailingAPIServicePolicy() *FailingAPIServicesPolicy {
	obj := &FailingAPIServicesPolicy{}
	s.CanIgnoreFailingAPIService = obj.CanIgnore
	return obj
}
