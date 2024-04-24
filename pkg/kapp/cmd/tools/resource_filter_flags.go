// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"fmt"
	"time"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type ResourceFilterFlags struct {
	Age string
	Rf  ctlres.ResourceFilter
	Bf  string
}

func (s *ResourceFilterFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.Age, "filter-age", "", "Set age filter (example: 5m-, 500h+, 10m-)")

	cmd.Flags().StringSliceVar(&s.Rf.Kinds, "filter-kind", nil, "Set kinds filter (example: Pod) (can repeat)")
	cmd.Flags().StringSliceVar(&s.Rf.Namespaces, "filter-ns", nil, "Set namespace filter (example: knative-serving) (can repeat)")
	cmd.Flags().StringSliceVar(&s.Rf.Names, "filter-name", nil, "Set name filter (example: controller) (can repeat)")
	cmd.Flags().StringSliceVar(&s.Rf.KindNames, "filter-kind-name", nil, "Set kind-name filter (example: Pod/controller) (can repeat)")
	cmd.Flags().StringSliceVar(&s.Rf.KindNamespaces, "filter-kind-ns", nil, "Set kind-namespace filter (example: Pod/, Pod/knative-serving) (can repeat)")
	cmd.Flags().StringSliceVar(&s.Rf.KindNsNames, "filter-kind-ns-name", nil, "Set kind-namespace-name filter (example: Deployment/knative-serving/controller) (can repeat)")
	cmd.Flags().StringSliceVar(&s.Rf.Labels, "filter-labels", nil, "Set label filter (example: x=y)")

	cmd.Flags().StringVar(&s.Bf, "filter", "", `Set filter (example: {"and":[{"not":{"resource":{"kinds":["foo%"]}}},{"resource":{"kinds":["!foo"]}}]})`)
}

func (s *ResourceFilterFlags) ResourceFilter() (ctlres.ResourceFilter, error) {
	createdAtBeforeTime, createdAtAfterTime, err := s.Times()
	if err != nil {
		return ctlres.ResourceFilter{}, err
	}

	rf := s.Rf
	rf.CreatedAtAfterTime = createdAtAfterTime
	rf.CreatedAtBeforeTime = createdAtBeforeTime

	if len(s.Bf) > 0 {
		boolFilter, err := ctlres.NewBoolFilterFromString(s.Bf)
		if err != nil {
			return ctlres.ResourceFilter{}, err
		}

		rf.BoolFilter = boolFilter
	}

	return rf, nil
}

func (s *ResourceFilterFlags) Times() (*time.Time, *time.Time, error) {
	if len(s.Age) == 0 {
		return nil, nil, nil
	}

	var ageStr string
	var ageOlder bool

	lastIdx := len(s.Age) - 1

	switch string(s.Age[lastIdx]) {
	case "+":
		ageStr = s.Age[:lastIdx]
		ageOlder = true
	case "-":
		ageStr = s.Age[:lastIdx]
	}

	dur, err := time.ParseDuration(ageStr)
	if err == nil {
		t1 := time.Now().UTC().Add(-dur)
		if ageOlder {
			return &t1, nil, nil
		}
		return nil, &t1, nil
	}

	return nil, nil, fmt.Errorf("Expected age filter to be either empty or " +
		"parseable time.Duration (example: 5m+, 24h-; valid units: ns, us, ms, s, m, h)")
}
