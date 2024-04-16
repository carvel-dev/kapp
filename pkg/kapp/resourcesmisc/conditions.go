// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type Conditions struct {
	resource ctlres.Resource
}

func (c Conditions) IsSelectedTrue(checkedTypes []string) (bool, string) {
	statuses := c.statuses()

	for _, t := range checkedTypes {
		status, found := statuses[t]
		if !found {
			return false, fmt.Sprintf("Condition %s is not set", t)
		}
		if status != "True" {
			return false, fmt.Sprintf("Condition %s is not True (%s)", t, status)
		}
	}

	return true, ""
}

func (c Conditions) IsAllTrue() (bool, string) {
	statuses := c.statuses()
	if len(statuses) == 0 {
		return false, "No conditions found"
	}

	for t, status := range c.statuses() {
		if status != "True" {
			return false, fmt.Sprintf("Condition %s is not True (%s)", t, status)
		}
	}

	return true, ""
}

func (c Conditions) statuses() map[string]string {
	statuses := map[string]string{}
	if conditions, ok := c.resource.Status()["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if typedCond, ok := cond.(map[string]interface{}); ok {
				if typedType, ok := typedCond["type"].(string); ok {
					if typedStatus, ok := typedCond["status"].(string); ok {
						statuses[typedType] = typedStatus
					}
				}
			}
		}
	}
	return statuses
}
