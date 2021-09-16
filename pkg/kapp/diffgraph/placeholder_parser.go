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

type PlaceholderParser struct {
	values map[string]string
}

func NewPlaceholderParserWithResource(resource ctlres.Resource) (PlaceholderParser, error) {
	values, err := valueMapFromResource(resource)
	if err != nil {
		return PlaceholderParser{}, fmt.Errorf("Could not obtain placeholder values from resource: %s", err)
	}
	return PlaceholderParser{values}, nil
}

func NewPlaceholderParserWithMap(values map[string]string) PlaceholderParser {
	return PlaceholderParser{values}
}

var (
	placeholderMatcher = regexp.MustCompile("{.+?}")
)

func (p PlaceholderParser) Parse(val string) (string, error) {
	placeholders := placeholderMatcher.FindAllString(val, -1)

	for _, placeholder := range placeholders {
		value, found := p.values[placeholder]
		if !found {
			return val, fmt.Errorf(`Expected placeholder to be one of these: %s but was %s`, p.placeholders(), placeholder)
		}
		if value == "" {
			return val, fmt.Errorf("Placeholder %s does not have a value for target resource, placeholders with the 'crd-' prefix can only be used with CRDs", placeholder)
		}
		val = strings.Replace(val, placeholder, value, 1)
	}

	return val, nil
}

func (p PlaceholderParser) placeholders() (placeholders []string) {
	for k := range p.values {
		placeholders = append(placeholders, k)
	}
	return placeholders
}

func valueMapFromResource(resource ctlres.Resource) (values map[string]string, err error) {
	var crdKind, crdGroup string
	crd := ctlcrd.NewAPIExtensionsVxCRD(resource)
	if crd != nil {
		crdKind, err = crd.Kind()
		if err != nil {
			return values, err
		}
		crdGroup, err = crd.Group()
		if err != nil {
			return values, err
		}
	}

	values = map[string]string{
		"{name}":      resource.Name(),
		"{crd-kind}":  crdKind,
		"{crd-group}": crdGroup,
		"{namespace}": resource.Namespace(),
		"{api-group}": resource.APIGroup(),
		"{kind}":      resource.Kind(),
	}

	return values, nil
}
