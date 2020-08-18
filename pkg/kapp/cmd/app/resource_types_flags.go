// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceTypesFlags struct {
	IgnoreFailingAPIServices   bool
	CanIgnoreFailingAPIService func(schema.GroupVersion) bool
}

func (s *ResourceTypesFlags) Set(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&s.IgnoreFailingAPIServices, "dangerous-ignore-failing-api-services",
		false, "Allow to ignore failing APIServices")
}

func (s *ResourceTypesFlags) FailingAPIServicePolicy() *FailingAPIServicesPolicy {
	obj := &FailingAPIServicesPolicy{}
	s.CanIgnoreFailingAPIService = obj.CanIgnore
	return obj
}
