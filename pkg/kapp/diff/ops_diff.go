// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"crypto/fips140"
	"crypto/md5"
	"fmt"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

type OpsDiff patch.Ops

func (l OpsDiff) HasChanges() bool { return len(l) > 0 }

// MinimalMD5 is a non-security convenience hash used to key/dedup diffs; it
// is not used for authentication or integrity verification.
// WithoutEnforcement lets it run under GODEBUG=fips140=only, which otherwise
// panics on any non-approved primitive regardless of how it's used.
func (l OpsDiff) MinimalMD5() string {
	var sum [md5.Size]byte
	fips140.WithoutEnforcement(func() {
		sum = md5.Sum([]byte(l.MinimalString()))
	})
	return fmt.Sprintf("%x", sum)
}

func (l OpsDiff) FullString() string { return "" }

func (l OpsDiff) MinimalString() string {
	opsDefs, err := patch.NewOpDefinitionsFromOps(patch.Ops(l))
	if err != nil {
		panic("building opdefs") // TODO panic
	}

	bs, err := yaml.Marshal(opsDefs)
	if err != nil {
		panic("yamling opsdiff") // TODO panic
	}

	return string(bs)
}
