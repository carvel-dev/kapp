package config

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type stringMapAppendRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
	SkipIfNotFound   bool
}

func (r stringMapAppendRule) AsMods(kvs map[string]string) []ctlres.StringMapAppendMod {
	var mods []ctlres.StringMapAppendMod

	for _, matcher := range r.ResourceMatchers {
		mods = append(mods, r.singleMod(matcher, kvs))
	}

	return mods
}

func (r stringMapAppendRule) singleMod(matcher ResourceMatcher, kvs map[string]string) ctlres.StringMapAppendMod {
	mod := ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AllMatcher{},
		Path:            r.Path,
		SkipIfNotFound:  r.SkipIfNotFound,
		KVs:             kvs,
	}

	switch {
	case matcher.KindNamespaceNameMatcher != nil:
		mod.ResourceMatcher = ctlres.KindNamespaceNameMatcher{
			Kind:      matcher.KindNamespaceNameMatcher.Kind,
			Namespace: matcher.KindNamespaceNameMatcher.Namespace,
			Name:      matcher.KindNamespaceNameMatcher.Name,
		}

	case matcher.APIVersionKindMatcher != nil:
		mod.ResourceMatcher = ctlres.APIVersionKindMatcher{
			APIVersion: matcher.APIVersionKindMatcher.APIVersion,
			Kind:       matcher.APIVersionKindMatcher.Kind,
		}

	default:
		mod.ResourceMatcher = ctlres.AllMatcher{}
	}

	return mod
}
