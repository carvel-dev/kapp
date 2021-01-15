// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"

	uitable "github.com/cppforlife/go-cli-ui/ui/table"
)

type ConditionsValue struct {
	status map[string]interface{}
}

func NewConditionsValue(status map[string]interface{}) ConditionsValue {
	return ConditionsValue{status}
}

func (t ConditionsValue) NeedsAttention() bool {
	if conditions, ok := t.status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if typedCond, ok := cond.(map[string]interface{}); ok {
				if typedStatus, ok := typedCond["status"].(string); ok {
					if typedStatus != "True" {
						return true
					}
				}
			}
		}
	}
	return false
}

func (t ConditionsValue) String() string {
	var totalNum, okNum int
	if conditions, ok := t.status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if typedCond, ok := cond.(map[string]interface{}); ok {
				if typedStatus, ok := typedCond["status"].(string); ok {
					totalNum++
					if typedStatus == "True" {
						okNum++
					}
				}
			}
		}
		return fmt.Sprintf("%d/%d t", okNum, totalNum)
	}
	return ""
}

func (t ConditionsValue) Value() uitable.Value { return t }

func (t ConditionsValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
