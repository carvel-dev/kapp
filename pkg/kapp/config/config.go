/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

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
	WaitRules           []WaitRule
	OwnershipLabelRules []OwnershipLabelRule
	LabelScopingRules   []LabelScopingRule
	TemplateRules       []TemplateRule
	DiffMaskRules       []DiffMaskRule

	AdditionalLabels                          map[string]string
	DiffAgainstLastAppliedFieldExclusionRules []DiffAgainstLastAppliedFieldExclusionRule

	// TODO additional?
	// TODO validations
	ChangeGroupBindings []ChangeGroupBinding
	ChangeRuleBindings  []ChangeRuleBinding
}

type WaitRule struct {
	SupportsObservedGeneration bool
	ConditionMatchers          []WaitRuleConditionMatcher
	ResourceMatchers           []ResourceMatcher
}

type WaitRuleConditionMatcher struct {
	Type    string
	Status  string
	Failure bool
	Success bool
}

type RebaseRule struct {
	ResourceMatchers []ResourceMatcher
	Path             ctlres.Path
	Paths            []ctlres.Path
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

type ChangeGroupBinding struct {
	Name             string
	ResourceMatchers []ResourceMatcher
}

type ChangeRuleBinding struct {
	Rules            []string
	IgnoreIfCyclical bool
	ResourceMatchers []ResourceMatcher
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

	for i, rule := range c.RebaseRules {
		err := rule.Validate()
		if err != nil {
			return fmt.Errorf("Validating rebase rule %d: %s", i, err)
		}
	}

	return nil
}

func (r RebaseRule) Validate() error {
	if len(r.Path) > 0 && len(r.Paths) > 0 {
		return fmt.Errorf("Expected only one of path or paths specified")
	}
	if len(r.Path) == 0 && len(r.Paths) == 0 {
		return fmt.Errorf("Expected either path or paths to be specified")
	}
	return nil
}

func (r RebaseRule) AsMods() []ctlres.ResourceModWithMultiple {
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
