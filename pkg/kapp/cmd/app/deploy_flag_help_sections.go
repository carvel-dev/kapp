// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/cppforlife/cobrautil"
	"github.com/spf13/cobra"
)

var (
	CommonFlagGroup = cobrautil.FlagHelpSection{
		Title:      "Common Flags:",
		ExactMatch: []string{"namespace", "app", "file", "diff-changes"},
	}
	DiffFlagGroup = cobrautil.FlagHelpSection{
		Title:       "Diff Flags:",
		PrefixMatch: "diff",
	}
	ApplyFlagGroup = cobrautil.FlagHelpSection{
		Title:       "Apply Flags:",
		PrefixMatch: "apply",
		ExactMatch: []string{
			"dangerous-allow-empty-list-of-resources",
			"dangerous-override-ownership-of-existing-resources",
		},
	}
	WaitFlagGroup = cobrautil.FlagHelpSection{
		Title:       "Wait Flags:",
		PrefixMatch: "wait",
		ExactMatch:  []string{"wait"},
	}
	ResourceFilterFlagGroup = cobrautil.FlagHelpSection{
		Title:       "Resource Filter Flags:",
		PrefixMatch: "filter",
		ExactMatch:  []string{"filter"},
	}
	ResourceValidationFlagGroup = cobrautil.FlagHelpSection{
		Title:       "Resource Validation Flags:",
		PrefixMatch: "allow",
	}
	ResourceManglingFlagGroup = cobrautil.FlagHelpSection{
		Title:      "Resource Mangling Flags:",
		ExactMatch: []string{"into-ns", "map-ns"},
	}
	LogsFlagGroup = cobrautil.FlagHelpSection{
		Title:       "Logs Flags:",
		PrefixMatch: "logs",
		ExactMatch:  []string{"logs"},
	}
	OtherFlagGroup = cobrautil.FlagHelpSection{
		Title:     "Available/Other Flags:",
		NoneMatch: true,
	}
)

func setDeployCmdFlags(cmd *cobra.Command) {
	cmd.SetUsageTemplate(cobrautil.FlagHelpSectionsUsageTemplate([]cobrautil.FlagHelpSection{
		CommonFlagGroup,
		DiffFlagGroup,
		ApplyFlagGroup,
		WaitFlagGroup,
		ResourceFilterFlagGroup,
		ResourceValidationFlagGroup,
		ResourceManglingFlagGroup,
		LogsFlagGroup,
		OtherFlagGroup,
	}))
}
