// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"strings"

	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
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

type HashSuffixResource struct {
	res ctlres.Resource
}

func (d HashSuffixResource) Res() ctlres.Resource {
	return d.res
}

func (d HashSuffixResource) SetBaseName(ver int) {
	// not necessary
}

func (d HashSuffixResource) BaseNameAndVersion() (string, string) {
	pieces := strings.Split(d.res.Name(), "-")
	if len(pieces) > 1 {
		return strings.Join(pieces[0:len(pieces)-1], "-"), ""
	}
	panic("expected suffix!")
}

func (d HashSuffixResource) Version() int {
	return -1
}

func (d HashSuffixResource) UniqVersionedKey() ctlres.UniqueResourceKey {
	baseName, _ := d.BaseNameAndVersion()
	return ctlres.NewUniqueResourceKeyWithCustomName(d.res, baseName)
}

func (d HashSuffixResource) UpdateAffected(rs []ctlres.Resource) error {
	// no updates required
	return nil
}
