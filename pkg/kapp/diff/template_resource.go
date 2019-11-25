package diff

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
)

const (
	nameSuffixSep = "-ver-"
)

type TemplateResource struct {
	res      ctlres.Resource
	allRules []ctlconf.TemplateRule
}

func (d TemplateResource) SetTemplatedName(ver int) {
	name := fmt.Sprintf("%s%s%d", d.res.Name(), nameSuffixSep, ver)
	d.res.SetName(name)
}

func (d TemplateResource) NonTemplatedName() (string, string) {
	name := d.res.Name()
	pieces := strings.Split(name, nameSuffixSep)
	if len(pieces) > 1 {
		return strings.Join(pieces[0:len(pieces)-1], nameSuffixSep), pieces[len(pieces)-1]
	}
	return name, ""
}

func (d TemplateResource) Version() int {
	_, ver := d.NonTemplatedName()
	if len(ver) == 0 {
		panic("Missing template version")
	}

	verInt, err1 := strconv.Atoi(ver)
	if err1 != nil {
		panic("Invalid template version")
	}

	return verInt
}

func (d TemplateResource) UniqTemplateKey() ctlres.UniqueResourceKey {
	nonTemplatedName, _ := d.NonTemplatedName()
	return ctlres.NewUniqueResourceKeyWithCustomName(d.res, nonTemplatedName)
}

func (d TemplateResource) UpdateAffected(rs []ctlres.Resource) error {
	rules, err := d.matchingRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		// TODO template that apply to other templates?
		err = d.updateAffected(rule, rs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d TemplateResource) updateAffected(rule ctlconf.TemplateRule, rs []ctlres.Resource) error {
	for _, affectedObjRef := range rule.AffectedResources.ObjectReferences {
		matchers := ctlconf.ResourceMatchers(affectedObjRef.ResourceMatchers).AsResourceMatchers()

		mod := ctlres.ObjectRefSetMod{
			ResourceMatcher: ctlres.AnyMatcher{matchers},
			Path:            affectedObjRef.Path,
			ReplacementFunc: d.buildObjRefReplacementFunc(affectedObjRef),
		}

		for _, res := range rs {
			err := mod.Apply(res)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d TemplateResource) buildObjRefReplacementFunc(
	affectedObjRef ctlconf.TemplateAffectedObjRef) func(map[string]interface{}) error {

	nonTemplatedName, _ := d.NonTemplatedName()

	return func(typedObj map[string]interface{}) error {
		bs, err := json.Marshal(typedObj)
		if err != nil {
			return fmt.Errorf("Remarshaling object reference: %s", err)
		}

		var objRef corev1.ObjectReference

		err = json.Unmarshal(bs, &objRef)
		if err != nil {
			return fmt.Errorf("Unmarshaling object reference: %s", err)
		}

		// Check as many rules as possible
		if len(affectedObjRef.NameKey) > 0 {
			if typedObj[affectedObjRef.NameKey] != nonTemplatedName {
				return nil
			}
		} else {
			if objRef.Name != nonTemplatedName {
				return nil
			}
		}

		if len(objRef.Namespace) > 0 && objRef.Namespace != d.res.Namespace() {
			return nil
		}
		if len(objRef.Kind) > 0 && objRef.Kind != d.res.Kind() {
			return nil
		}
		if len(objRef.APIVersion) > 0 && objRef.APIVersion != d.res.APIVersion() {
			return nil
		}

		if len(affectedObjRef.NameKey) > 0 {
			typedObj[affectedObjRef.NameKey] = d.res.Name()
		} else {
			typedObj["name"] = d.res.Name()
		}

		return nil
	}
}

func (d TemplateResource) matchingRules() ([]ctlconf.TemplateRule, error) {
	var result []ctlconf.TemplateRule

	for _, rule := range d.allRules {
		matchers := ctlconf.ResourceMatchers(rule.ResourceMatchers).AsResourceMatchers()
		if (ctlres.AnyMatcher{matchers}).Matches(d.res) {
			result = append(result, rule)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Expected to find at least one template rule")
	}

	return result, nil
}
