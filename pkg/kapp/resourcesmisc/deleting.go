// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type Deleting struct {
	resource ctlres.Resource
}

func NewDeleting(resource ctlres.Resource) *Deleting {
	if resource.IsDeleting() {
		return &Deleting{resource}
	}
	return nil
}

func (s Deleting) IsDoneApplying() DoneApplyState {
	return DoneApplyState{Done: false, Message: "Deleting"}
}
