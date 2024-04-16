// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"
	"strings"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
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
	if len(s.resource.Finalizers()) > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf("Waiting on finalizers: %s",
			strings.Join(s.resource.Finalizers(), ", "))}
	}
	return DoneApplyState{Done: false, Message: "Deleting"}
}
