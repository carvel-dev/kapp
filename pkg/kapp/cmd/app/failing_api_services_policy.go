// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type FailingAPIServicesPolicy struct {
	rs         []ctlres.Resource
	gvs        []schema.GroupVersion
	configured bool
}

func (s *FailingAPIServicesPolicy) MarkRequiredResources(rs []ctlres.Resource) {
	s.rs = rs
	s.configured = true
}

func (s *FailingAPIServicesPolicy) MarkRequiredGVs(gvs []schema.GroupVersion) {
	s.gvs = gvs
	s.configured = true
}

func (s *FailingAPIServicesPolicy) CanIgnore(groupVer schema.GroupVersion) bool {
	if !s.configured {
		return false
	}
	groupVerStr := groupVer.String()
	for _, res := range s.rs {
		if res.APIVersion() == groupVerStr {
			return false
		}
	}
	for _, gv := range s.gvs {
		if gv.String() == groupVerStr {
			return false
		}
	}
	return true
}

func (s *FailingAPIServicesPolicy) GVs(rs1 []ctlres.Resource, rs2 []ctlres.Resource) []schema.GroupVersion {
	var result []schema.GroupVersion
	for _, rs := range [][]ctlres.Resource{rs1, rs2} {
		for _, res := range rs {
			result = append(result, res.GroupVersion())
		}
	}
	return result
}
