package resourcesmisc

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type BatchvxCronJob struct {
	resource ctlres.Resource
}

func NewBatchvxCronJob(resource ctlres.Resource) *BatchvxCronJob {
	matcher := ctlres.APIGroupKindMatcher{
		APIGroup: "batch",
		Kind:     "CronJob",
	}
	if matcher.Matches(resource) {
		return &BatchvxCronJob{resource}
	}
	return nil
}

func (s BatchvxCronJob) IsDoneApplying() DoneApplyState {
	// Always return success as we do not want to pick up associated
	// pods that might have previously failed
	return DoneApplyState{Done: true, Successful: true}
}
