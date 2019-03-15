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
					totalNum += 1
					if typedStatus == "True" {
						okNum += 1
					}
				}
			}
		}
		return fmt.Sprintf("%d OK / %d", okNum, totalNum)
	}
	return ""
}

func (t ConditionsValue) Value() uitable.Value { return t }

func (t ConditionsValue) Compare(other uitable.Value) int {
	panic("Not implemented")
}
