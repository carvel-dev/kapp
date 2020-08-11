// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type LabelFlags struct {
	Labels []string
}

func (s *LabelFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.Labels, "labels", nil, "Set app label (format: key=val, key=) (can repeat)")
}

func (s *LabelFlags) AsMap() (map[string]string, error) {
	result := map[string]string{}
	for _, val := range s.Labels {
		pieces := strings.SplitN(val, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Expected label to be in 'key=val' format")
		}
		if len(pieces[0]) == 0 {
			return nil, fmt.Errorf("Expected label key to be non-empty")
		}
		result[pieces[0]] = pieces[1]
	}
	return result, nil
}
