// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"carvel.dev/kapp/pkg/kapp/version"
	"carvel.dev/kapp/pkg/kapp/yttresmod"
	semver "github.com/hashicorp/go-version"
	"sigs.k8s.io/yaml"
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
	WaitRules           []WaitRule
	OwnershipLabelRules []OwnershipLabelRule
	LabelScopingRules   []LabelScopingRule
	TemplateRules       []TemplateRule
	DiffMaskRules       []DiffMaskRule
	PreflightRules      []PreflightRule

	AdditionalLabels                          map[string]string
	DiffAgainstLastAppliedFieldExclusionRules []DiffAgainstLastAppliedFieldExclusionRule
	DiffAgainstExistingFieldExclusionRules    []DiffAgainstExistingFieldExclusionRule

	// TODO additional?
	// TODO validations
	ChangeGroupBindings []ChangeGroupBinding
	ChangeRuleBindings  []ChangeRuleBinding
}

type WaitRule struct {
	SupportsObservedGeneration bool
	ConditionMatchers          []WaitRuleConditionMatcher
	ResourceMatchers           []ResourceMatcher
	Ytt                        *WaitRuleYtt
}

type WaitRuleConditionMatcher struct {
	Type                       string
	Status                     string
	Failure                    bool
	Success                    bool
	SupportsObservedGeneration bool
	UnblockChanges             bool
}

type WaitRuleYtt struct {
	// Contracts are named and versioned (eg v1)
	// to provide a stable interface to rule authors.
	// Multiple contracts will be offered at the same time
	// so that existing rules do not not break as we decide to evolve running environment.
	FuncContractV1 *FuncContractV1 `json:"funcContractV1"`
}

type FuncContractV1 struct {
	Resource string `json:"resource.star"`
}

type RebaseRule struct {
	ResourceMatchers []ResourceMatcher

	Path    ctlres.Path
	Paths   []ctlres.Path
	Type    string
	Sources []ctlres.FieldCopyModSource

	Ytt *RebaseRuleYtt
}

type RebaseRuleYtt struct {
	// Contracts are named (eg overlay) and versioned (eg v1)
	// to provide a stable interface to rule authors.
	// Multiple contracts will be offered at the same time
	// so that existing rules do not not break as we decide to evolve running environment.
	OverlayContractV1 *RebaseRuleYttOverlayContractV1 `json:"overlayContractV1"`
}

type RebaseRuleYttOverlayContractV1 struct {
	OverlayYAML string `json:"overlay.yml"`
}

type DiffAgainstLastAppliedFieldExclusionRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
}

type DiffAgainstExistingFieldExclusionRule struct {
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
	IsDefault        bool `json:"isDefault"`
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

type ChangeGroupBinding struct {
	Name             string
	ResourceMatchers []ResourceMatcher
}

type ChangeRuleBinding struct {
	Rules            []string
	IgnoreIfCyclical bool
	ResourceMatchers []ResourceMatcher
}

type PreflightRule struct {
	Name   string
	Config map[string]any
}

func NewConfigFromResource(res ctlres.Resource) (Config, error) {
	if res.APIVersion() != configAPIVersion {
		return Config{}, fmt.Errorf(
			"Expected kapp config to have apiVersion '%s', but was '%s'",
			configAPIVersion, res.APIVersion())
	}

	if res.Kind() != configKind {
		return Config{}, fmt.Errorf(
			"Expected kapp config to have kind '%s', but was '%s'",
			configKind, res.Kind())
	}

	bs, err := res.AsYAMLBytes()
	if err != nil {
		return Config{}, err
	}

	return newConfigFromYAMLBytes(bs, res.Description())
}

func newConfigFromYAMLBytes(bs []byte, description string) (Config, error) {
	var config Config
	err := yaml.Unmarshal(bs, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Unmarshaling %s: %w", description, err)
	}

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("Validating config: %w", err)
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
			return fmt.Errorf("Parsing minimum version constraint: %w", err)
		}

		kappVersion, err := semver.NewVersion(version.Version)
		if err != nil {
			return fmt.Errorf("Parsing version constraint: %w", err)
		}

		if !userConstraint.Check(kappVersion) {
			return fmt.Errorf("kapp version '%s' does "+
				"not meet the minimum required version '%s'", version.Version, c.MinimumRequiredVersion)
		}
	}

	for i, rule := range c.RebaseRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("Validating rebase rule %d: %w", i, err)
		}
	}

	return nil
}

func (r RebaseRule) Validate() error {
	if r.Ytt != nil {
		if len(r.Path) > 0 || len(r.Paths) > 0 || len(r.Type) > 0 || len(r.Sources) > 0 {
			return fmt.Errorf("Expected only resourceMatchers specified with ytt configuration")
		}
		return nil
	}
	if len(r.Path) > 0 && len(r.Paths) > 0 {
		return fmt.Errorf("Expected only one of path or paths specified")
	}
	if len(r.Path) == 0 && len(r.Paths) == 0 {
		return fmt.Errorf("Expected either path or paths to be specified")
	}
	return nil
}

func (r RebaseRule) AsMods() []ctlres.ResourceModWithMultiple {
	if r.Ytt != nil {
		switch {
		case r.Ytt.OverlayContractV1 != nil:
			return []ctlres.ResourceModWithMultiple{yttresmod.OverlayContractV1Mod{
				ResourceMatcher: ctlres.AnyMatcher{
					Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
				},
				OverlayYAML: r.Ytt.OverlayContractV1.OverlayYAML,
			}}

		default:
			panic("Unknown rebase rule ytt contract (supported: overlayContractV1)")
		}
	}

	var mods []ctlres.ResourceModWithMultiple
	var paths []ctlres.Path

	if len(r.Paths) == 0 {
		paths = append(paths, r.Path)
	} else {
		paths = r.Paths
	}

	for _, path := range paths {
		switch r.Type {
		case "copy":
			mods = append(mods, ctlres.FieldCopyMod{
				ResourceMatcher: ctlres.AnyMatcher{
					Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
				},
				Path:    path,
				Sources: r.Sources,
			})

		case "remove":
			mods = append(mods, ctlres.FieldRemoveMod{
				ResourceMatcher: ctlres.AnyMatcher{
					Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
				},
				Path: path,
			})

		default:
			panic(fmt.Sprintf("Unknown rebase rule type: %s (supported: copy, remove)", r.Type)) // TODO
		}
	}

	return mods
}

func (r DiffAgainstLastAppliedFieldExclusionRule) AsMod() ctlres.FieldRemoveMod {
	return ctlres.FieldRemoveMod{
		ResourceMatcher: ctlres.AnyMatcher{
			Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
		},
		Path: r.Path,
	}
}
func (r DiffAgainstExistingFieldExclusionRule) AsMod() ctlres.FieldRemoveMod {
	return ctlres.FieldRemoveMod{
		ResourceMatcher: ctlres.AnyMatcher{
			Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
		},
		Path: r.Path,
	}
}

func (r OwnershipLabelRule) AsMod(kvs map[string]string) ctlres.StringMapAppendMod {
	return ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AnyMatcher{
			Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
		},
		Path: r.Path,
		KVs:  kvs,
	}
}

func (r LabelScopingRule) AsMod(kvs map[string]string) ctlres.StringMapAppendMod {
	return ctlres.StringMapAppendMod{
		ResourceMatcher: ctlres.AnyMatcher{
			Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
		},
		Path:           r.Path,
		SkipIfNotFound: true,
		KVs:            kvs,
	}
}

func (r WaitRule) ResourceMatcher() ctlres.ResourceMatcher {
	return ctlres.AnyMatcher{
		Matchers: ResourceMatchers(r.ResourceMatchers).AsResourceMatchers(),
	}
}
