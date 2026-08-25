// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"crypto/fips140"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/k14s/difflib"
)

type TextDiff struct {
	recs []difflib.DiffRecord
}

func NewTextDiff(existingLines, newLines []string, allowAnchoredDiff bool) TextDiff {
	if allowAnchoredDiff && (len(existingLines) > 500 || len(newLines) > 500) {
		// Diff is memory hungry, use AnchoredDiff for large resources
		return TextDiff{difflib.AnchoredDiff(existingLines, newLines)}
	}
	return TextDiff{difflib.Diff(existingLines, newLines)}
}

func (l TextDiff) Records() []difflib.DiffRecord { return l.recs }

func (l TextDiff) HasChanges() bool {
	for _, diff := range l.recs {
		if diff.Delta != difflib.Common {
			return true
		}
	}
	return false
}

// MinimalMD5 is a non-security convenience hash used to key/dedup diffs; it
// is not used for authentication or integrity verification.
// WithoutEnforcement lets it run under GODEBUG=fips140=only, which otherwise
// panics on any non-approved primitive regardless of how it's used.
func (l TextDiff) MinimalMD5() string {
	var sum [md5.Size]byte
	fips140.WithoutEnforcement(func() {
		sum = md5.Sum([]byte(l.MinimalString()))
	})
	return fmt.Sprintf("%x", sum)
}

func (l TextDiff) MinimalString() string { return l.String(false) }
func (l TextDiff) FullString() string    { return l.String(true) }

func (l TextDiff) String(full bool) string {
	var sb strings.Builder

	for _, diff := range l.recs {
		var mark string

		switch diff.Delta {
		case difflib.RightOnly:
			mark = " + "
		case difflib.LeftOnly:
			mark = " - "
		case difflib.Common:
			if !full {
				continue
			}
			mark = "   "
		}

		// make sure to have line numbers to make sure diff is truly unique
		sb.WriteString(fmt.Sprintf("%3d,%3d%s%s\n", diff.LineLeft, diff.LineRight, mark, diff.Payload))
	}

	return sb.String()
}
