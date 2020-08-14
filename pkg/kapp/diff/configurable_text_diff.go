// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"strings"

	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ConfigurableTextDiff struct {
	existingRes, newRes ctlres.Resource
	ignored             bool

	memoizedTextDiff *TextDiff
}

func NewConfigurableTextDiff(existingRes, newRes ctlres.Resource, ignored bool) *ConfigurableTextDiff {
	return &ConfigurableTextDiff{existingRes, newRes, ignored, nil}
}

func (d ConfigurableTextDiff) Full() TextDiff {
	if d.memoizedTextDiff == nil {
		textDiff := d.calculate(d.existingRes, d.newRes)
		d.memoizedTextDiff = &textDiff
	}
	return *d.memoizedTextDiff
}

func (d ConfigurableTextDiff) Masked(rules []ctlconf.DiffMaskRule) (TextDiff, error) {
	var existingRes, newRes ctlres.Resource
	var err error

	if d.existingRes != nil {
		existingRes, err = NewMaskedResource(d.existingRes, rules).Resource()
		if err != nil {
			return TextDiff{}, fmt.Errorf("Masking existing resource: %s", err)
		}
	}

	if d.newRes != nil {
		newRes, err = NewMaskedResource(d.newRes, rules).Resource()
		if err != nil {
			return TextDiff{}, fmt.Errorf("Masking new resource: %s", err)
		}
	}

	return d.calculate(existingRes, newRes), nil
}

func (d ConfigurableTextDiff) calculate(existingRes, newRes ctlres.Resource) TextDiff {
	existingLines := []string{}
	newLines := []string{}

	if existingRes != nil {
		existingBytes, err := existingRes.AsYAMLBytes()
		if err != nil {
			panic("yamling existingRes") // TODO panic
		}
		existingLines = strings.Split(string(existingBytes), "\n")
	}

	if newRes != nil {
		newBytes, err := newRes.AsYAMLBytes()
		if err != nil {
			panic("yamling newRes") // TODO panic
		}
		newLines = strings.Split(string(newBytes), "\n")
	} else if d.ignored {
		newLines = existingLines // show as no changes
	}

	return NewTextDiff(existingLines, newLines)
}
