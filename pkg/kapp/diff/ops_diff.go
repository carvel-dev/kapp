// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"crypto/md5"
	"fmt"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

type OpsDiff patch.Ops

func (l OpsDiff) HasChanges() bool { return len(l) > 0 }

func (l OpsDiff) MinimalMD5() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(l.MinimalString())))
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
