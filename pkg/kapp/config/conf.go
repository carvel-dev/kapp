package config

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type Conf struct {
	configs []Config
}

func NewConfFromResources(resources []ctlres.Resource) ([]ctlres.Resource, Conf, error) {
	var rsWithoutConfigs []ctlres.Resource
	var configs []Config

	for _, res := range resources {
		if res.APIVersion() == configAPIVersion {
			if res.Kind() == configKind {
				config, err := NewConfigFromResource(res)
				if err != nil {
					return nil, Conf{}, err
				}
				configs = append(configs, config)
			} else {
				errMsg := "Unexpected kind in resource '%s', wanted '%s'"
				return nil, Conf{}, fmt.Errorf(errMsg, res.Description(), configKind)
			}
		} else {
			rsWithoutConfigs = append(rsWithoutConfigs, res)
		}
	}

	cfg := Conf{configs}

	ctlres.SetGlobalWaitingRules(cfg.WaitRuleMods())

	return rsWithoutConfigs, cfg, nil
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
			mods = append(mods, rule.AsMods()...)
		}
	}
	return mods
}

func (c Conf) OwnershipLabelMods() func(kvs map[string]string) []ctlres.StringMapAppendMod {
	return func(kvs map[string]string) []ctlres.StringMapAppendMod {
		var mods []ctlres.StringMapAppendMod
		for _, config := range c.configs {
			for _, rule := range config.OwnershipLabelRules {
				mods = append(mods, rule.AsMods(kvs)...)
			}
		}
		return mods
	}
}

func (c Conf) WaitRuleMods() []ctlres.WaitingRuleMod {
	var mods []ctlres.WaitingRuleMod // not config.WaitingRule
	for _, cfg := range c.configs {
		for _, rule := range cfg.WaitingRules {
			mods = append(mods, ctlres.WaitingRuleMod{
				SupportsObservedGeneration: rule.SupportsObservedGeneration,
				SuccessfulConditions:       rule.SuccessfulConditions,
				FailureConditions:          rule.FailureConditions,
				ResourceMatchers:           rule.ResourceMatchers.AsResourceMatchers(),
			})
		}
	}
	return mods
}

func (c Conf) LabelScopingMods() func(kvs map[string]string) []ctlres.StringMapAppendMod {
	return func(kvs map[string]string) []ctlres.StringMapAppendMod {
		var mods []ctlres.StringMapAppendMod
		for _, config := range c.configs {
			for _, rule := range config.LabelScopingRules {
				mods = append(mods, rule.AsMods(kvs)...)
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
