// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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

type VersionedResource struct {
	res                 ctlres.Resource
	allRules            []ctlconf.TemplateRule
	stripNameHashSuffix bool
}

func (d VersionedResource) StripNameHashSuffix() bool {

	stripEnabled := d.stripNameHashSuffix

	// TODO should properly check using a ResourceMatcher
	matchesRequiredKind := d.res.Kind() == "ConfigMap" || d.res.Kind() == "Secret"

	return stripEnabled && matchesRequiredKind

}

func (d VersionedResource) SetBaseName(ver int) {
	if d.trackVersions() {
		currentName, _ := d.BaseNameAndVersion()
		name := fmt.Sprintf("%s%s%d", currentName, nameSuffixSep, ver)
		d.res.SetName(name)
	}
}

func (d VersionedResource) BaseNameAndVersion() (string, string) {
	name := d.res.Name()
	pieces := strings.Split(name, nameSuffixSep)
	if len(pieces) > 1 {
		return strings.Join(pieces[0:len(pieces)-1], nameSuffixSep), pieces[len(pieces)-1]
	}
	if d.StripNameHashSuffix() {
		pieces = strings.Split(name, "-")
		if len(pieces) > 1 {
			return strings.Join(pieces[0:len(pieces)-1], "-"), ""
		}
	}
	return name, ""
}

func (d VersionedResource) Version() int {
	_, ver := d.BaseNameAndVersion()
	if len(ver) == 0 {
		if d.StripNameHashSuffix() {
			return -1
		}
		panic(fmt.Sprintf("Missing version in versioned resource '%s'", d.res.Description()))
	}

	verInt, err1 := strconv.Atoi(ver)
	if err1 != nil {
		panic(fmt.Sprintf("Invalid version in versioned resource '%s'", d.res.Description()))
	}

	return verInt
}

func (d VersionedResource) UniqVersionedKey() ctlres.UniqueResourceKey {
	baseName, _ := d.BaseNameAndVersion()
	return ctlres.NewUniqueResourceKeyWithCustomName(d.res, baseName)
}

func (d VersionedResource) trackVersions() bool {
	_, hasVersionedAnn := d.res.Annotations()[versionedResAnnKey]
	return hasVersionedAnn
}

func (d VersionedResource) IsVersioned() bool {
	if d.StripNameHashSuffix() {
		return true
	}
	return d.trackVersions()
}

func (d VersionedResource) IsTracked() bool {
	return d.IsVersioned() && d.trackVersions()
}

func (d VersionedResource) IsExistingVersioned() bool {

	// Expect that versioned resources should not be transient
	// (Annotations may have been copied from versioned resources
	// onto transient resources for non-versioning related purposes).
	notTransient := !d.res.Transient()

	// TODO check moved during refactoring, but not sure why it is required
	_, version := d.BaseNameAndVersion()
	hasVersion := version != ""

	versionUnnecessary := !d.trackVersions()

	return d.IsVersioned() && notTransient && (hasVersion || versionUnnecessary)

}

func (d VersionedResource) AssignNextVersion(old VersionedResource) {
	d.SetBaseName(old.Version() + 1)
}

func (d VersionedResource) AssignNewVersion() {
	d.SetBaseName(1)
}

func (d VersionedResource) UpdateAffected(rs []ctlres.Resource) error {

	if !d.trackVersions() {
		// if we're not tracking versions we don't update any names and thus
		// don't need to update any references.
		return nil
	}

	rules, err := d.matchingRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		// TODO versioned resources that affect other versioned resources
		err = d.updateAffected(rule, rs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d VersionedResource) updateAffected(rule ctlconf.TemplateRule, rs []ctlres.Resource) error {
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

	for _, res := range rs {
		for k, v := range res.Annotations() {
			if k == explicitReferenceKey || strings.HasPrefix(k, explicitReferenceKeyPrefix) {
				explicitRef := NewExplicitVersionedRef(k, v)
				objectRef, err := explicitRef.AsObjectRef()
				if err != nil {
					return fmt.Errorf("Parsing versioned explicit ref on resource '%s': %s", res.Description(), err)
				}

				// Passing empty TemplateAffectedObjRef as explicit references do not have a special name key
				err = d.buildObjRefReplacementFunc(ctlconf.TemplateAffectedObjRef{})(objectRef)
				if err != nil {
					return fmt.Errorf("Processing object ref for explicit ref on resource '%s': %s", res.Description(), err)
				}

				annotationMod, err := explicitRef.AnnotationMod(objectRef)
				if err != nil {
					return fmt.Errorf("Preparing annotation mod for versioned explicit ref on resource '%s': %s", res.Description(), err)
				}

				err = annotationMod.Apply(res)
				if err != nil {
					return fmt.Errorf("Updating versioned explicit ref on resource '%s': %s", res.Description(), err)
				}
			}
		}
	}

	return nil
}

func (d VersionedResource) buildObjRefReplacementFunc(
	affectedObjRef ctlconf.TemplateAffectedObjRef) func(map[string]interface{}) error {

	baseName, _ := d.BaseNameAndVersion()

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
			if typedObj[affectedObjRef.NameKey] != baseName {
				return nil
			}
		} else {
			if objRef.Name != baseName {
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

func (d VersionedResource) matchingRules() ([]ctlconf.TemplateRule, error) {
	var result []ctlconf.TemplateRule

	for _, rule := range d.allRules {
		matchers := ctlconf.ResourceMatchers(rule.ResourceMatchers).AsResourceMatchers()
		if (ctlres.AnyMatcher{matchers}).Matches(d.res) {
			result = append(result, rule)
		}
	}

	return result, nil
}
