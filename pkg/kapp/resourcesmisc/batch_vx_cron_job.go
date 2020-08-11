// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type BatchVxCronJob struct {
	resource ctlres.Resource
}

func NewBatchVxCronJob(resource ctlres.Resource) *BatchVxCronJob {
	matcher := ctlres.APIGroupKindMatcher{
		APIGroup: "batch",
		Kind:     "CronJob",
	}
	if matcher.Matches(resource) {
		return &BatchVxCronJob{resource}
	}
	return nil
}

func (s BatchVxCronJob) IsDoneApplying() DoneApplyState {
	// Always return success as we do not want to pick up associated
	// pods that might have previously failed
	return DoneApplyState{Done: true, Successful: true}
}
