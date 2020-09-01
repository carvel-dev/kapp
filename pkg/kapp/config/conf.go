// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
)

const (
	configLabelKey     = "kapp.k14s.io/config"
	configMapConfigKey = "config.yml"
)

type Conf struct {
	configs []Config
}

func NewConfFromResources(resources []ctlres.Resource) ([]ctlres.Resource, Conf, error) {
	var rsWithoutConfigs []ctlres.Resource
	var configs []Config

	for _, res := range resources {
		_, isLabeledAsConfig := res.Labels()[configLabelKey]

		switch {
		case res.APIVersion() == configAPIVersion:
			config, err := NewConfigFromResource(res)
			if err != nil {
				return nil, Conf{}, fmt.Errorf(
					"Parsing resource '%s' as kapp config: %s", res.Description(), err)
			}
			configs = append(configs, config)

		case isLabeledAsConfig:
			config, err := newConfigFromConfigMapRes(res)
			if err != nil {
				return nil, Conf{}, fmt.Errorf(
					"Parsing resource '%s' labeled as kapp config: %s", res.Description(), err)
			}
			// Make sure to add ConfigMap resource to regular resources list
			// (our goal of allowing kapp config in ConfigMaps is to allow
			// both kubectl and kapp to work against exactly same configuration;
			// hence want to preserve same behaviour)
			rsWithoutConfigs = append(rsWithoutConfigs, res)
			configs = append(configs, config)

		default:
			rsWithoutConfigs = append(rsWithoutConfigs, res)
		}
	}

	return rsWithoutConfigs, Conf{configs}, nil
}

func newConfigFromConfigMapRes(res ctlres.Resource) (Config, error) {
	if res.APIVersion() != "v1" || res.Kind() != "ConfigMap" {
		errMsg := "Expected kapp config to be within v1/ConfigMap but apiVersion or kind do not match"
		return Config{}, fmt.Errorf(errMsg, res.Description())
	}

	configCM := corev1.ConfigMap{}

	err := res.AsTypedObj(&configCM)
	if err != nil {
		return Config{}, fmt.Errorf("Converting resource to ConfigMap: %s", err)
	}

	configStr, found := configCM.Data[configMapConfigKey]
	if !found {
		return Config{}, fmt.Errorf("Expected to find field 'data.\"%s\"', but did not", configMapConfigKey)
	}

	configRes, err := ctlres.NewResourceFromBytes([]byte(configStr))
	if err != nil {
		return Config{}, fmt.Errorf("Parsing kapp config as resource: %s", err)
	}

	return NewConfigFromResource(configRes)
}

func (c Conf) RebaseMods() []ctlres.ResourceModWithMultiple {
	var mods []ctlres.ResourceModWithMultiple
	for _, config := range c.configs {
		for _, rule := range config.RebaseRules {
			mods = append(mods, rule.AsMods()...)
		}
	}
	return mods
}

func (c Conf) DiffAgainstLastAppliedFieldExclusionMods() []ctlres.FieldRemoveMod {
	var mods []ctlres.FieldRemoveMod
	for _, config := range c.configs {
		for _, rule := range config.DiffAgainstLastAppliedFieldExclusionRules {
			mods = append(mods, rule.AsMod())
		}
	}
	return mods
}

func (c Conf) OwnershipLabelMods() func(kvs map[string]string) []ctlres.StringMapAppendMod {
	return func(kvs map[string]string) []ctlres.StringMapAppendMod {
		var mods []ctlres.StringMapAppendMod
		for _, config := range c.configs {
			for _, rule := range config.OwnershipLabelRules {
				mods = append(mods, rule.AsMod(kvs))
			}
		}
		return mods
	}
}

func (c Conf) WaitRules() []WaitRule {
	var rules []WaitRule
	for _, config := range c.configs {
		rules = append(rules, config.WaitRules...)
	}
	return rules
}

func (c Conf) LabelScopingMods() func(kvs map[string]string) []ctlres.StringMapAppendMod {
	return func(kvs map[string]string) []ctlres.StringMapAppendMod {
		var mods []ctlres.StringMapAppendMod
		for _, config := range c.configs {
			for _, rule := range config.LabelScopingRules {
				mods = append(mods, rule.AsMod(kvs))
			}
		}
		return mods
	}
}

func (c Conf) TemplateRules() []TemplateRule {
	var result []TemplateRule
	for _, config := range c.configs {
		result = append(result, config.TemplateRules...)
	}
	return result
}

func (c Conf) DiffMaskRules() []DiffMaskRule {
	var result []DiffMaskRule
	for _, config := range c.configs {
		result = append(result, config.DiffMaskRules...)
	}
	return result
}

func (c Conf) AdditionalLabels() map[string]string {
	result := map[string]string{}
	for _, config := range c.configs {
		for k, v := range config.AdditionalLabels {
			result[k] = v
		}
	}
	return result
}

func (c Conf) ChangeGroupBindings() []ChangeGroupBinding {
	var result []ChangeGroupBinding
	for _, config := range c.configs {
		result = append(result, config.ChangeGroupBindings...)
	}
	return result
}

func (c Conf) ChangeRuleBindings() []ChangeRuleBinding {
	var result []ChangeRuleBinding
	for _, config := range c.configs {
		result = append(result, config.ChangeRuleBindings...)
	}
	return result
}
