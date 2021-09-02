// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

type ExistsOpts struct {
	SameUID bool
}

func (e *ExistsOpts) checkForSameUID() bool {
	return e.SameUID
}

