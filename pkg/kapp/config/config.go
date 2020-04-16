package config

import (
	"fmt"

	"github.com/ghodss/yaml"
	semver "github.com/hashicorp/go-version"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/k14s/kapp/pkg/kapp/version"
)

const (
	configAPIVersion = "kapp.k14s.io/v1alpha1"
	configKind       = "Config"
)

type Config struct {
	APIVersion string `json:"apiVersion"`
	Kind       string

	MinimumRequiredVersion string `json:"minimumRequiredVersion,omitempty"`

	RebaseRules         []RebaseRule
	OwnershipLabelRules []OwnershipLabelRule
	LabelScopingRules   []LabelScopingRule
	TemplateRules       []TemplateRule
	DiffMaskRules       []DiffMaskRule

	AdditionalLabels                          map[string]string
	DiffAgainstLastAppliedFieldExclusionRules []DiffAgainstLastAppliedFieldExclusionRule
}

type RebaseRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
	Type             string
	Sources          []ctlres.FieldCopyModSource
}

type DiffAgainstLastAppliedFieldExclusionRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
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

type DiffMaskRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
}

type TemplateAffectedResources struct {
	ObjectReferences []TemplateAffectedObjRef
	// TODO support label injections?
}

type TemplateAffectedObjRef struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
	NameKey          string `json:"nameKey"`
}

type ResourceMatchers []ResourceMatcher

type ResourceMatcher struct {
	AllMatcher               *AllMatcher // default
	AnyMatcher               *AnyMatcher
	APIVersionKindMatcher    *APIVersionKindMatcher `json:"apiVersionKindMatcher"`
	KindNamespaceNameMatcher *KindNamespaceNameMatcher
}

type AllMatcher struct{}

type AnyMatcher struct {
	Matchers []ResourceMatcher
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

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("Validating config: %s", err)
	}

	return config, nil
}

func (c Config) Validate() error {
	if c.APIVersion != configAPIVersion {
		return fmt.Errorf("Validating apiVersion: Unknown version (known: %s)", configAPIVersion)
	}
	if c.Kind != configKind {
		return fmt.Errorf("Validating kind: Unknown kind (known: %s)", configKind)
	}

	if len(c.MinimumRequiredVersion) > 0 {
		if c.MinimumRequiredVersion[0] == 'v' {
			return fmt.Errorf("Validating minimum version: Must not have prefix 'v' (e.g. '0.8.0')")
		}

		userConstraint, err := semver.NewConstraint(">=" + c.MinimumRequiredVersion)
		if err != nil {
			return fmt.Errorf("Parsing minimum version constraint: %s", err)
		}

		kappVersion, err := semver.NewVersion(version.Version)
		if err != nil {
			return fmt.Errorf("Parsing version constraint: %s", err)
		}

		if !userConstraint.Check(kappVersion) {
			return fmt.Errorf("kapp version '%s' does "+
				"not meet the minimum required version '%s'", version.Version, c.MinimumRequiredVersion)
		}
	}

	return nil
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

func (r DiffAgainstLastAppliedFieldExclusionRule) AsMods() []ctlres.FieldRemoveMod {
	var mods []ctlres.FieldRemoveMod
	for _, matcher := range r.ResourceMatchers {
		mods = append(mods, ctlres.FieldRemoveMod{
			ResourceMatcher: matcher.AsResourceMatcher(),
			Path:            r.Path,
		})
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
	case m.AllMatcher != nil:
		return ctlres.AllResourceMatcher{}

	case m.AnyMatcher != nil:
		return ctlres.AnyResourceMatcher{
			Matchers: ResourceMatchers(m.AnyMatcher.Matchers).AsResourceMatchers(),
		}

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
		panic("Unknown resource matcher specified")
	}
}
