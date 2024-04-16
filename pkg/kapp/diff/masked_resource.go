// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"encoding/json"
	"fmt"
	"sort"

	ctlconf "carvel.dev/kapp/pkg/kapp/config"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type MaskedResource struct {
	res   ctlres.Resource
	rules []ctlconf.DiffMaskRule
}

func NewMaskedResource(res ctlres.Resource, rules []ctlconf.DiffMaskRule) MaskedResource {
	if res == nil {
		panic("Expected res be non-nil")
	}
	return MaskedResource{res, rules}
}

func (r MaskedResource) Resource() (ctlres.Resource, error) {
	res := r.res.DeepCopy()
	for _, rule := range r.rules {
		err := r.update(rule, res)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (r MaskedResource) update(rule ctlconf.DiffMaskRule, res ctlres.Resource) error {
	mod := ctlres.ObjectRefSetMod{
		ResourceMatcher: ctlres.AnyMatcher{
			ctlconf.ResourceMatchers(rule.ResourceMatchers).AsResourceMatchers(),
		},
		Path:            rule.Path,
		ReplacementFunc: r.maskValues,
	}
	return mod.Apply(res)
}

var (
	maskedResourceValues       = map[string]int{}
	maskedResourceValueLastIdx = 1
)

func (MaskedResource) maskValues(typedObj map[string]interface{}) error {
	// Needed for deterministic value indexing
	var sortedKeys []string
	for k := range typedObj {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		var maskVal string

		valBs, err := json.Marshal(typedObj[k])
		if err != nil {
			// Prefer to be a unique value in diff, even if actual value is the same
			// since it's better to indicate change when there is no change vs
			// no change when there is actually a change.
			maskVal = fmt.Sprintf("<-- unknown value not shown (#%d)", maskedResourceValueLastIdx)
			maskedResourceValueLastIdx++
		} else {
			valIdx, found := maskedResourceValues[string(valBs)]
			if !found {
				valIdx = maskedResourceValueLastIdx
				maskedResourceValues[string(valBs)] = valIdx
				maskedResourceValueLastIdx++
			}
			maskVal = fmt.Sprintf("<-- value not shown (#%d)", valIdx)
		}

		typedObj[k] = maskVal
	}
	return nil
}
