package resources

type ResourceMatcher interface {
	Matches(Resource) bool
}

type AnyMatcher struct {
	Matchers []ResourceMatcher
}

var _ ResourceMatcher = AnyMatcher{}

func (m AnyMatcher) Matches(res Resource) bool {
	for _, matcher := range m.Matchers {
		if matcher.Matches(res) {
			return true
		}
	}
	return false
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

type AllResourceMatcher struct{}

var _ ResourceMatcher = AllResourceMatcher{}

func (AllResourceMatcher) Matches(Resource) bool { return true }

type AnyResourceMatcher struct {
	Matchers []ResourceMatcher
}

var _ ResourceMatcher = AnyResourceMatcher{}

func (m AnyResourceMatcher) Matches(res Resource) bool {
	for _, m := range m.Matchers {
		if m.Matches(res) {
			return true
		}
	}
	return false
}

type NotResourceMatcher struct {
	Matcher ResourceMatcher
}

var _ ResourceMatcher = NotResourceMatcher{}

func (m NotResourceMatcher) Matches(res Resource) bool {
	return !m.Matcher.Matches(res)
}

type AndResourceMatcher struct {
	Matchers []ResourceMatcher
}

var _ ResourceMatcher = AndResourceMatcher{}

func (m AndResourceMatcher) Matches(res Resource) bool {
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
