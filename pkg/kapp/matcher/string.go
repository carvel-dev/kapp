// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package matcher

import (
	"regexp"
	"strings"
)

const (
	stringMatcherGlob1 = '*'
	stringMatcherGlob2 = '%' // not special char in Bash
)

type StringMatcher struct {
	expected string
}

func NewStringMatcher(expected string) StringMatcher {
	return StringMatcher{expected}
}

func (f StringMatcher) Matches(actual string) bool {
	firstChar := f.expected[0]
	lastChar := f.expected[len(f.expected)-1]

	prefixGlob := firstChar == stringMatcherGlob1 || firstChar == stringMatcherGlob2
	suffixGlob := lastChar == stringMatcherGlob1 || lastChar == stringMatcherGlob2

	switch {
	case prefixGlob && suffixGlob:
		return regexp.MustCompile(regexp.QuoteMeta(f.expected[1 : len(f.expected)-1])).MatchString(actual)

	case prefixGlob:
		return strings.HasSuffix(actual, f.expected[1:])

	case suffixGlob:
		return strings.HasPrefix(actual, f.expected[:len(f.expected)-1])

	default:
		return actual == f.expected
	}
}
