// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type stripNameHashSuffixConfig struct {
	Enabled         bool
	ResourceMatcher ctlres.ResourceMatcher
}

func (d stripNameHashSuffixConfig) EnabledFor(res ctlres.Resource) (stripEnabled bool) {
	if d.Enabled {
		return d.ResourceMatcher.Matches(res)
	}
	return false
}

func newStripNameHashSuffixConfig(enabled bool, resourceMatchers [][]ctlres.ResourceMatcher) (result stripNameHashSuffixConfig) {
	result = stripNameHashSuffixConfig{Enabled: enabled}
	if enabled {
		var allMatchers []ctlres.ResourceMatcher
		for _, matchers := range resourceMatchers {
			allMatchers = append(allMatchers, ctlres.AndMatcher{
				Matchers: matchers,
			})
		}
		result.ResourceMatcher = ctlres.AndMatcher{Matchers: allMatchers}
	}
	return result
}

func newStripNameHashSuffixConfigFromConf(confs ctlconf.StripNameHashSuffixConfigs) stripNameHashSuffixConfig {
	enabled, matchers := confs.AggregateToCtlRes()
	return newStripNameHashSuffixConfig(enabled, matchers)
}
