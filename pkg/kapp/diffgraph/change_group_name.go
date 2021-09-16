// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
	"regexp"
	"strings"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlcrd "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
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
	placeholders := placeholderMatcher.FindAllString(c.name, -1)
	if len(placeholders) < 1 {
		return c.name, nil
	}

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

	for _, placeholder := range placeholders {
		value, found := values[placeholder]
		if !found {
			return c.name, fmt.Errorf("Expected placeholder to be one of these: %s but was %s", c.placeholders(values), placeholder)
		}
		if value == "" {
			return c.name, fmt.Errorf("Placeholder %s does not have a value for target resource (hint: placeholders with the 'crd-' prefix can only be used with CRDs)", placeholder)
		}
		c.name = strings.Replace(c.name, placeholder, value, 1)
	}

	return c.name, nil
}

func (c ChangeGroupName) placeholders(values map[string]string) (placeholders []string) {
	for k := range values {
		placeholders = append(placeholders, k)
	}
	return placeholders
}
