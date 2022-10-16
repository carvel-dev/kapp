// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"strings"

	ctlconf "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/config"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
)

type stripNameHashSuffixConfig struct {
	ResourceMatcher ctlres.ResourceMatcher
}

func (d stripNameHashSuffixConfig) EnabledFor(res ctlres.Resource) (stripEnabled bool) {
	return d.ResourceMatcher.Matches(res)
}

func newStripNameHashSuffixConfigFromConf(confs []ctlconf.StripNameHashSuffixConfig) stripNameHashSuffixConfig {
	includeMatchers := []ctlres.ResourceMatcher{}
	excludeMatchers := []ctlres.ResourceMatcher{}
	for _, conf := range confs {
		includeMatchers = append(includeMatchers, conf.IncludeMatchers()...)
		excludeMatchers = append(excludeMatchers, conf.ExcludeMatchers()...)
	}
	return stripNameHashSuffixConfig{
		ResourceMatcher: ctlres.AndMatcher{
			Matchers: []ctlres.ResourceMatcher{
				ctlres.AnyMatcher{Matchers: includeMatchers},
				ctlres.NotMatcher{
					Matcher: ctlres.AnyMatcher{Matchers: excludeMatchers},
				},
			},
		},
	}
}

func newStripNameHashSuffixConfigEmpty() stripNameHashSuffixConfig {
	return newStripNameHashSuffixConfigFromConf([]ctlconf.StripNameHashSuffixConfig{})
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
