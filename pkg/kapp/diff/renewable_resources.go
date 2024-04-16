// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"time"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

const (
	renewDurationAnnKey = "kapp.k14s.io/renew-duration"
	lastRenewTimeAnnKey = "kapp.k14s.io/last-renewed-time"
)

type RenewableResources struct {
	existingRs, newRs []ctlres.Resource
}

func NewRenewableResources(existingRs, newRs []ctlres.Resource) *RenewableResources {
	return &RenewableResources{existingRs: existingRs, newRs: newRs}
}

func (d RenewableResources) Prepare() error {
	exResourcesMap := existingResourcesMap(d.existingRs)

	for _, res := range d.newRs {
		val, found := res.Annotations()[renewDurationAnnKey]
		if found {
			duration, err := time.ParseDuration(val)
			if err != nil {
				return fmt.Errorf("Unable to parse renew-duration for resource %s: %s", res.Description(), err.Error())
			}

			resKey := ctlres.NewUniqueResourceKey(res).String()
			exRes := exResourcesMap[resKey]
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

func (d RenewableResources) addLastRenewedTimeAnn(res, exRes ctlres.Resource, duration time.Duration) error {
	var (
		lastRenewed time.Time
		err         error
	)

	val, found := exRes.Annotations()[lastRenewTimeAnnKey]
	if found {
		lastRenewed, err = time.Parse(time.RFC3339, val)
		if err != nil {
			return fmt.Errorf("Unable to parse last-renewed-time for resource %s: %s", res.Description(), err.Error())
		}
	}

	if lastRenewed.Before(exRes.CreatedAt()) {
		lastRenewed = exRes.CreatedAt()
	}

	renewTime := fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339))
	if time.Now().Before(lastRenewed.Add(duration)) {
		if !found {
			return nil
		}
		renewTime = val
	}

	return ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs: map[string]string{
			lastRenewTimeAnnKey: renewTime,
		},
	}.Apply(res)
}
