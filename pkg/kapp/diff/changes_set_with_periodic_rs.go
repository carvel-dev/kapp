// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"time"

	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

type ChangeSetWithPeriodicRes struct {
	existingRs, newRs []ctlres.Resource
}

func NewChangeSetWithPeriodicRs(existingRs, newRs []ctlres.Resource) *ChangeSetWithPeriodicRes {
	return &ChangeSetWithPeriodicRes{existingRs: existingRs, newRs: newRs}
}

func (d ChangeSetWithPeriodicRes) Calculate() error {
	existingRs := existingVersionedResources(d.existingRs)
	existingVersionRsGrouped := newGroupedVersionedResources(existingRs.Versioned).groupedRes

	return d.checkAndUpdateResWithMaxDurationAnn(existingVersionRsGrouped, existingRs)
}

func (d ChangeSetWithPeriodicRes) checkAndUpdateResWithMaxDurationAnn(existingVersionRsGrouped map[string][]ctlres.Resource,
	existingRs versionedResources) error {

	exNonVersionResMap := map[string]ctlres.Resource{}

	for _, res := range existingRs.NonVersioned {
		resKey := ctlres.NewUniqueResourceKey(res).String()
		exNonVersionResMap[resKey] = res
	}

	for _, res := range d.newRs {
		if val, found := res.Annotations()[maxDurationAnnKey]; found {
			duration, err := time.ParseDuration(val)
			if err != nil {
				return fmt.Errorf("Unable to parse max-duration for resource %s: %s", res.Description(), err.Error())
			}

			resKey := ctlres.NewUniqueResourceKey(res).String()
			exRes, found := exNonVersionResMap[resKey]
			if !found {
				if exResGroup, found := existingVersionRsGrouped[resKey]; found {
					exRes = exResGroup[len(exResGroup)-1]
				}
			}

			if exRes != nil {
				err = d.addLastRenewedTimeAnn(res, exRes, duration)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d ChangeSetWithPeriodicRes) addLastRenewedTimeAnn(res, exRes ctlres.Resource, duration time.Duration) error {
	var (
		lastRenewed time.Time
		err         error
	)

	if val, found := exRes.Annotations()[lastRenewTimeAnnKey]; found {
		lastRenewed, err = time.Parse(time.RFC3339, val)
		if err != nil {
			return fmt.Errorf("Unable to parse last-renewed-time for resource %s: %s", res.Description(), err.Error())
		}
	}

	if lastRenewed.Before(exRes.CreatedAt()) {
		lastRenewed = exRes.CreatedAt()
	}

	if time.Now().Before(lastRenewed.Add(duration)) {
		return nil
	}

	return ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs: map[string]string{
			lastRenewTimeAnnKey: fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339)),
		},
	}.Apply(res)
}
