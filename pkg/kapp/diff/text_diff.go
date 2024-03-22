// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
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

func (l TextDiff) MinimalMD5() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(l.MinimalString())))
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
