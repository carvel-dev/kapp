// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type SimpleLabel struct {
	labelSelector labels.Selector
}

func NewSimpleLabel(labelSelector labels.Selector) SimpleLabel {
	return SimpleLabel{labelSelector}
}

func (a SimpleLabel) KV() (string, string, error) {
	reqs, selectable := a.labelSelector.Requirements()
	if !selectable {
		return "", "", fmt.Errorf("Expected label selector to be selectable")
	}

	if len(reqs) != 1 {
		return "", "", fmt.Errorf("Expected label selector to have one label KV")
	}

	key := reqs[0].Key()
	op := reqs[0].Operator()
	val, _ := reqs[0].Values().PopAny()

	if op != selection.Equals && op != selection.In {
		return "", "", fmt.Errorf("Expected label selector to check for equality")
	}

	if reqs[0].Values().Len() != 1 {
		return "", "", fmt.Errorf("Expected label selector to check against single label value")
	}

	return key, val, nil
}
