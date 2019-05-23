package config

import (
	"fmt"

	"github.com/ghodss/yaml"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const (
	configAPIVersion = "kapp.k14s.io/v1alpha1"
	configKind       = "Config"
)

type Config struct {
	APIVersion          string `json:"apiVersion"`
	Kind                string
	RebaseRules         []RebaseRule
	OwnershipLabelRules []OwnershipLabelRule
	LabelScopingRules   []LabelScopingRule
	TemplateRules       []TemplateRule
}

type RebaseRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
	Type             string
	Sources          []ctlres.FieldCopyModSource
}

type OwnershipLabelRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
}

type LabelScopingRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
}

type TemplateRule struct {
	ResourceMatchers  []ResourceMatcher
	AffectedResources TemplateAffectedResources
}

type TemplateAffectedResources struct {
	ObjectReferences []TemplateAffectedObjRef
	// TODO support label injections?
}

type TemplateAffectedObjRef struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
}

type ResourceMatchers []ResourceMatcher

type ResourceMatcher struct {
	AllResourceMatcher       *AllResourceMatcher    // default
	APIVersionKindMatcher    *APIVersionKindMatcher `json:"apiVersionKindMatcher"`
	KindNamespaceNameMatcher *KindNamespaceNameMatcher
}

type AllResourceMatcher struct{}

type APIVersionKindMatcher struct {
	APIVersion string `json:"apiVersion"`
	Kind       string
}

type KindNamespaceNameMatcher struct {
	Kind      string
	Namespace string
	Name      string
}

func NewConfigFromResource(res ctlres.Resource) (Config, error) {
	bs, err := res.AsYAMLBytes()
	if err != nil {
		return Config{}, err
	}

	var config Config

	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Unmarshaling %s: %s", res.Description(), err)
	}

	return config, nil
}

func (r RebaseRule) AsMods() []ctlres.ResourceModWithMultiple {
	var mods []ctlres.ResourceModWithMultiple

	for _, matcher := range r.ResourceMatchers {
		switch r.Type {
		case "copy":
			mods = append(mods, ctlres.FieldCopyMod{
				ResourceMatcher: matcher.AsResourceMatcher(),
				Path:            r.Path,
				Sources:         r.Sources,
			})

		case "remove":
			mods = append(mods, ctlres.FieldRemoveMod{
				ResourceMatcher: matcher.AsResourceMatcher(),
				Path:            r.Path,
			})

		default:
			panic(fmt.Sprintf("Unknown rebase rule type: %s (supported: copy, remove)", r.Type)) // TODO
		}
	}

	return mods
}

func (r OwnershipLabelRule) AsMods(kvs map[string]string) []ctlres.StringMapAppendMod {
	return stringMapAppendRule{ResourceMatchers: r.ResourceMatchers, Path: r.Path}.AsMods(kvs)
}

func (r LabelScopingRule) AsMods(kvs map[string]string) []ctlres.StringMapAppendMod {
	return stringMapAppendRule{ResourceMatchers: r.ResourceMatchers, Path: r.Path, SkipIfNotFound: true}.AsMods(kvs)
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
	case m.KindNamespaceNameMatcher != nil:
		return ctlres.KindNamespaceNameMatcher{
			Kind:      m.KindNamespaceNameMatcher.Kind,
			Namespace: m.KindNamespaceNameMatcher.Namespace,
			Name:      m.KindNamespaceNameMatcher.Name,
		}

	case m.APIVersionKindMatcher != nil:
		return ctlres.APIVersionKindMatcher{
			APIVersion: m.APIVersionKindMatcher.APIVersion,
			Kind:       m.APIVersionKindMatcher.Kind,
		}

	default:
		return ctlres.AllResourceMatcher{}
	}
}
