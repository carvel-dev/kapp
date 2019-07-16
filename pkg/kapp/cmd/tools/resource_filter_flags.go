package tools

import (
	"fmt"
	"time"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/spf13/cobra"
)

type ResourceFilterFlags struct {
	age string
	rf  ctlres.ResourceFilter
	bf  string
}

func (s *ResourceFilterFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.age, "filter-age", "", "Set age filter (example: 5m)")

	cmd.Flags().StringSliceVar(&s.rf.Kinds, "filter-kind", nil, "Set kinds filter (example: Pod) (can repeat)")
	cmd.Flags().StringSliceVar(&s.rf.Namespaces, "filter-ns", nil, "Set namespace filter (example: knative-serving) (can repeat)")
	cmd.Flags().StringSliceVar(&s.rf.Names, "filter-name", nil, "Set name filter (example: controller) (can repeat)")
	cmd.Flags().StringSliceVar(&s.rf.KindNamespaces, "filter-kind-ns", nil, "Set kind-namespace filter (example: Pod/, Pod/knative-serving) (can repeat)")
	cmd.Flags().StringSliceVar(&s.rf.KindNsNames, "filter-kind-ns-name", nil, "Set kind-namespace-name filter (example: Deployment/knative-serving/controller) (can repeat)")

	cmd.Flags().StringVar(&s.bf, "filter", "", `Set filter (example: {"and":[{"not":{"match":{kind":"foo%"}}},{"kind":"!foo"}]})`)
}

func (s *ResourceFilterFlags) ResourceFilter() (ctlres.ResourceFilter, error) {
	createdAt, err := s.AfterTime()
	if err != nil {
		return ctlres.ResourceFilter{}, err
	}

	rf := s.rf
	rf.CreatedAtAfterTime = createdAt

	if len(s.bf) > 0 {
		boolFilter, err := ctlres.NewBoolFilterFromString(s.bf)
		if err != nil {
			return ctlres.ResourceFilter{}, err
		}

		rf.BoolFilter = boolFilter
	}

	return rf, nil
}

func (s *ResourceFilterFlags) AfterTime() (*time.Time, error) {
	if len(s.age) == 0 {
		return nil, nil
	}

	dur, err := time.ParseDuration(s.age)
	if err == nil {
		t1 := time.Now().UTC().Add(-dur)
		return &t1, nil
	}

	return nil, fmt.Errorf("Expected age filter to be either empty or parseable time.Duration (eg 5m)")
}
