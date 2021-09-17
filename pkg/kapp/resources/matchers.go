// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

type ResourceMatcher interface {
	Matches(Resource) bool
}

type APIGroupKindMatcher struct {
	APIGroup string
	Kind     string
}

var _ ResourceMatcher = APIGroupKindMatcher{}

func (m APIGroupKindMatcher) Matches(res Resource) bool {
	return res.APIGroup() == m.APIGroup && res.Kind() == m.Kind
}

type APIVersionKindMatcher struct {
	APIVersion string
	Kind       string
}

var _ ResourceMatcher = APIVersionKindMatcher{}

func (m APIVersionKindMatcher) Matches(res Resource) bool {
	return res.APIVersion() == m.APIVersion && res.Kind() == m.Kind
}

type KindNamespaceNameMatcher struct {
	Kind, Namespace, Name string
}

var _ ResourceMatcher = KindNamespaceNameMatcher{}

func (m KindNamespaceNameMatcher) Matches(res Resource) bool {
	return res.Kind() == m.Kind && res.Namespace() == m.Namespace && res.Name() == m.Name
}

type AllMatcher struct{}

var _ ResourceMatcher = AllMatcher{}

func (AllMatcher) Matches(Resource) bool { return true }

type AnyMatcher struct {
	Matchers []ResourceMatcher
}

var _ ResourceMatcher = AnyMatcher{}

func (m AnyMatcher) Matches(res Resource) bool {
	for _, m := range m.Matchers {
		if m.Matches(res) {
			return true
		}
	}
	return false
}

type NotMatcher struct {
	Matcher ResourceMatcher
}

var _ ResourceMatcher = NotMatcher{}

func (m NotMatcher) Matches(res Resource) bool {
	return !m.Matcher.Matches(res)
}

type AndMatcher struct {
	Matchers []ResourceMatcher
}

var _ ResourceMatcher = AndMatcher{}

func (m AndMatcher) Matches(res Resource) bool {
	for _, m := range m.Matchers {
		if !m.Matches(res) {
			return false
		}
	}
	return true
}

type HasAnnotationMatcher struct {
	Keys []string
}

var _ ResourceMatcher = HasAnnotationMatcher{}

func (m HasAnnotationMatcher) Matches(res Resource) bool {
	anns := res.Annotations()
	for _, key := range m.Keys {
		if _, found := anns[key]; !found {
			return false
		}
	}
	return true
}

type HasNamespaceMatcher struct {
	Names []string
}

var _ ResourceMatcher = HasNamespaceMatcher{}

func (m HasNamespaceMatcher) Matches(res Resource) bool {
	resNs := res.Namespace()
	if len(resNs) == 0 {
		return false // cluster resource
	}
	if len(m.Names) == 0 {
		return true // matches any name, but not cluster
	}
	for _, name := range m.Names {
		if name == resNs {
			return true
		}
	}
	return false
}

var (
	// TODO should we just generically match *.k8s.io?
	// Based on https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#-strong-api-groups-strong-
	builtinAPIGroups = map[string]struct{}{
		"":                             struct{}{},
		"admissionregistration.k8s.io": struct{}{},
		"apiextensions.k8s.io":         struct{}{},
		"apiregistration.k8s.io":       struct{}{},
		"apps":                         struct{}{},
		"authentication.k8s.io":        struct{}{},
		"authorization.k8s.io":         struct{}{},
		"autoscaling":                  struct{}{},
		"batch":                        struct{}{},
		"certificates.k8s.io":          struct{}{},
		"coordination.k8s.io":          struct{}{},
		"discovery.k8s.io":             struct{}{},
		"events.k8s.io":                struct{}{},
		"extensions":                   struct{}{},
		"flowcontrol.apiserver.k8s.io": struct{}{},
		"internal.apiserver.k8s.io":    struct{}{},
		"metrics.k8s.io":               struct{}{},
		"migration.k8s.io":             struct{}{},
		"networking.k8s.io":            struct{}{},
		"node.k8s.io":                  struct{}{},
		"policy":                       struct{}{},
		"rbac.authorization.k8s.io":    struct{}{},
		"scheduling.k8s.io":            struct{}{},
		"storage.k8s.io":               struct{}{},
	}
)

type CustomResourceMatcher struct{}

var _ ResourceMatcher = CustomResourceMatcher{}

func (m CustomResourceMatcher) Matches(res Resource) bool {
	_, found := builtinAPIGroups[res.APIGroup()]
	return !found
}
