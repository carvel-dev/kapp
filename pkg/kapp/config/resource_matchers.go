// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type ResourceMatchers []ResourceMatcher

type ResourceMatcher struct {
	AllMatcher               *AllMatcher // default
	AnyMatcher               *AnyMatcher
	NotMatcher               *NotMatcher
	AndMatcher               *AndMatcher
	APIGroupKindMatcher      *APIGroupKindMatcher
	APIVersionKindMatcher    *APIVersionKindMatcher `json:"apiVersionKindMatcher"`
	KindNamespaceNameMatcher *KindNamespaceNameMatcher
	HasAnnotationMatcher     *HasAnnotationMatcher
	HasNamespaceMatcher      *HasNamespaceMatcher
	CustomResourceMatcher    *CustomResourceMatcher
	EmptyFieldMatcher        *EmptyFieldMatcher
}

type AllMatcher struct{}

type AnyMatcher struct {
	Matchers []ResourceMatcher
}

type NotMatcher struct {
	Matcher ResourceMatcher
}

type AndMatcher struct {
	Matchers []ResourceMatcher
}

type APIGroupKindMatcher struct {
	APIGroup string `json:"apiGroup"`
	Kind     string
}

type APIVersionKindMatcher struct {
	APIVersion string `json:"apiVersion"`
	Kind       string
}

type KindNamespaceNameMatcher struct {
	Kind      string
	Namespace string
	Name      string
}

type HasAnnotationMatcher struct {
	Keys []string
}

type HasNamespaceMatcher struct {
	Names []string
}

type CustomResourceMatcher struct{}

type EmptyFieldMatcher struct {
	Path ctlres.Path
}

func (ms ResourceMatchers) AsResourceMatchers() []ctlres.ResourceMatcher {
	var result []ctlres.ResourceMatcher
	for _, matcher := range ms {
		result = append(result, matcher.AsResourceMatcher())
	}
	return result
}

func (m ResourceMatcher) AsResourceMatcher() ctlres.ResourceMatcher {
	switch {
	case m.AllMatcher != nil:
		return ctlres.AllMatcher{}

	case m.AnyMatcher != nil:
		return ctlres.AnyMatcher{
			Matchers: ResourceMatchers(m.AnyMatcher.Matchers).AsResourceMatchers(),
		}

	case m.AndMatcher != nil:
		return ctlres.AndMatcher{
			Matchers: ResourceMatchers(m.AndMatcher.Matchers).AsResourceMatchers(),
		}

	case m.NotMatcher != nil:
		return ctlres.NotMatcher{
			Matcher: m.NotMatcher.Matcher.AsResourceMatcher(),
		}

	case m.KindNamespaceNameMatcher != nil:
		return ctlres.KindNamespaceNameMatcher{
			Kind:      m.KindNamespaceNameMatcher.Kind,
			Namespace: m.KindNamespaceNameMatcher.Namespace,
			Name:      m.KindNamespaceNameMatcher.Name,
		}

	case m.APIGroupKindMatcher != nil:
		return ctlres.APIGroupKindMatcher{
			APIGroup: m.APIGroupKindMatcher.APIGroup,
			Kind:     m.APIGroupKindMatcher.Kind,
		}

	case m.APIVersionKindMatcher != nil:
		return ctlres.APIVersionKindMatcher{
			APIVersion: m.APIVersionKindMatcher.APIVersion,
			Kind:       m.APIVersionKindMatcher.Kind,
		}

	case m.HasAnnotationMatcher != nil:
		return ctlres.HasAnnotationMatcher{
			Keys: m.HasAnnotationMatcher.Keys,
		}

	case m.HasNamespaceMatcher != nil:
		return ctlres.HasNamespaceMatcher{
			Names: m.HasNamespaceMatcher.Names,
		}

	case m.CustomResourceMatcher != nil:
		return ctlres.CustomResourceMatcher{}

	case m.EmptyFieldMatcher != nil:
		return ctlres.EmptyFieldMatcher{Path: m.EmptyFieldMatcher.Path}

	default:
		panic(fmt.Sprintf("Unknown resource matcher specified: %#v", m))
	}
}
