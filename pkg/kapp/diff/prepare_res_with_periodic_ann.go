// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"strings"
	"time"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const (
	MaxDurationAnnKey   = "kapp.k14s.io/max-duration"
	lastRenewTimeAnnKey = "kapp.k14s.io/last-renewed-time"
)

func CheckAndCalculateResWithPeriodicAnn(existingResources, newResources []ctlres.Resource) error {
	var (
		duration time.Duration
		err      error
	)

	addNonceMod := ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            ctlres.NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs: map[string]string{
			lastRenewTimeAnnKey: fmt.Sprintf("%v", time.Now().UTC().Format(time.RFC3339)),
		},
	}
	existRes := existingPeriodicResource(existingResources)

	for _, res := range newResources {
		rs, found := existRes[res.Name()+res.Kind()]
		if found {
			val, found := rs.Annotations()[MaxDurationAnnKey]
			if !found {
				continue
			}

			duration, err = time.ParseDuration(val)
			if err != nil {
				return err
			}

			if found && time.Now().After(rs.CreatedAt().Add(duration)) {
				err = addNonceMod.Apply(res)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func existingPeriodicResource(existingResources []ctlres.Resource) map[string]ctlres.Resource {
	exRes := make(map[string]ctlres.Resource)
	for _, rs := range existingResources {
		if _, found := rs.Annotations()[MaxDurationAnnKey]; found {
			if _, found := rs.Annotations()[versionedResAnnKey]; found {
				name := strings.Split(rs.Name(), "-ver-")[0]
				exRes[name+rs.Kind()] = rs
			} else {
				exRes[rs.Name()+rs.Kind()] = rs
			}
		}
	}
	return exRes
}
