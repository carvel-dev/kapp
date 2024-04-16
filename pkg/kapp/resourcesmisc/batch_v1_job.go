// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type BatchV1Job struct {
	resource ctlres.Resource
}

func NewBatchV1Job(resource ctlres.Resource) *BatchV1Job {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "batch/v1",
		Kind:       "Job",
	}
	if matcher.Matches(resource) {
		return &BatchV1Job{resource}
	}
	return nil
}

func (s BatchV1Job) IsDoneApplying() DoneApplyState {
	job := batchv1.Job{}

	err := s.resource.AsTypedObj(&job)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false,
			Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	for _, cond := range job.Status.Conditions {
		switch {
		case cond.Type == batchv1.JobComplete && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: true, Successful: true, Message: "Completed"}

		case cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: true, Successful: false,
				Message: fmt.Sprintf("Failed with reason %s: %s", cond.Reason, cond.Message)}
		}
	}

	return DoneApplyState{Done: false, Message: fmt.Sprintf(
		"Waiting to complete (%d active, %d failed, %d succeeded)",
		job.Status.Active, job.Status.Failed, job.Status.Succeeded)}
}

/*

status:
  conditions:
  - lastProbeTime: "2019-06-26T22:18:22Z"
    lastTransitionTime: "2019-06-26T22:18:22Z"
    message: Job has reached the specified backoff limit
    reason: BackoffLimitExceeded
    status: "True"
    type: Failed
  failed: 7
  startTime: "2019-06-26T22:07:50Z"

*/
