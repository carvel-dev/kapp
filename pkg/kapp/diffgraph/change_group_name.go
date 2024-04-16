// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
	"regexp"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlcrd "carvel.dev/kapp/pkg/kapp/resourcesmisc"
)

type ChangeGroupName struct {
	name     string
	resource ctlres.Resource
}

func NewChangeGroupNameForResource(name string, resource ctlres.Resource) ChangeGroupName {
	return ChangeGroupName{name, resource}
}

var (
	placeholderMatcher = regexp.MustCompile("{.+?}")
)

// Placeholders have the format {placeholder-name}
// Other patterns like ${placeholder-name} are commonly used by other operators/tools
func (c ChangeGroupName) AsString() (string, error) {
	var crdKind, crdGroup string
	var err error
	crd := ctlcrd.NewAPIExtensionsVxCRD(c.resource)
	if crd != nil {
		crdKind, err = crd.Kind()
		if err != nil {
			return c.name, err
		}
		crdGroup, err = crd.Group()
		if err != nil {
			return c.name, err
		}
	}

	values := map[string]string{
		"{api-group}": c.resource.APIGroup(),
		"{kind}":      c.resource.Kind(),
		"{name}":      c.resource.Name(),
		"{namespace}": c.resource.Namespace(),
		"{crd-kind}":  crdKind,
		"{crd-group}": crdGroup,
	}

	replaced := placeholderMatcher.ReplaceAllStringFunc(c.name, func(placeholder string) string {
		value, found := values[placeholder]
		if !found {
			err = fmt.Errorf("Expected placeholder to be one of these: %s but was %s", c.placeholders(values), placeholder)
		}
		if value == "" {
			err = fmt.Errorf("Placeholder %s does not have a value for target resource (hint: placeholders with the 'crd-' prefix can only be used with CRDs)", placeholder)
		}
		return value
	})

	return replaced, err
}

func (c ChangeGroupName) placeholders(values map[string]string) (placeholders []string) {
	for k := range values {
		placeholders = append(placeholders, k)
	}
	return placeholders
}
