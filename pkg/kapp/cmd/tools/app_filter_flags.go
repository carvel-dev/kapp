// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"fmt"
	"time"

	"carvel.dev/kapp/pkg/kapp/app"
	"github.com/spf13/cobra"
)

type AppFilterFlags struct {
	age    string
	labels []string

	af app.AppFilter
}

func (s *AppFilterFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.age, "filter-age", "", "Set age filter (example: 5m-, 500h+, 10m-)")
	cmd.Flags().StringSliceVar(&s.labels, "filter-labels", nil, "Set label filter (example: x=y)")
}

func (s *AppFilterFlags) AppFilter() (app.AppFilter, error) {
	createdAtBeforeTime, createdAtAfterTime, err := s.Times()
	if err != nil {
		return app.AppFilter{}, err
	}

	af := s.af
	af.CreatedAtAfterTime = createdAtAfterTime
	af.CreatedAtBeforeTime = createdAtBeforeTime

	return af, nil
}

func (s *AppFilterFlags) Labels() []string {
	return s.labels
}

func (s *AppFilterFlags) Times() (*time.Time, *time.Time, error) {
	if len(s.age) == 0 {
		return nil, nil, nil
	}

	var (
		ageStr   string
		ageOlder bool
	)

	lastIdx := len(s.age) - 1

	switch string(s.age[lastIdx]) {
	case "+":
		ageStr = s.age[:lastIdx]
		ageOlder = true
	case "-":
		ageStr = s.age[:lastIdx]
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
