// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"strings"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type LabeledApp struct {
	labelSelector       labels.Selector
	identifiedResources ctlres.IdentifiedResources
}

var _ App = &LabeledApp{}

func (a *LabeledApp) Name() string {
	str := a.labelSelector.String()
	if len(str) == 0 {
		return "?"
	}
	return str
}

func (a *LabeledApp) Namespace() string { return "" }

func (a *LabeledApp) Description() string {
	return fmt.Sprintf("labeled app '%s'", a.Name())
}

func (a *LabeledApp) LabelSelector() (labels.Selector, error) {
	return a.labelSelector, nil
}

func (a *LabeledApp) UsedGVs() ([]schema.GroupVersion, error)       { return nil, nil }
func (a *LabeledApp) UpdateUsedGVs(gvs []schema.GroupVersion) error { return nil }

func (a *LabeledApp) CreateOrUpdate(labels map[string]string) error { return nil }
func (a *LabeledApp) Exists() (bool, string, error)                 { return true, "", nil }

func (a *LabeledApp) Delete() error {
	labelSelector, err := a.LabelSelector()
	if err != nil {
		return err
	}

	rs, err := a.identifiedResources.List(labelSelector, nil)
	if err != nil {
		return fmt.Errorf("Relisting app resources: %s", err)
	}

	if len(rs) > 0 {
		var resourceNames []string
		for _, res := range rs {
			resourceNames = append(resourceNames, res.Description())
		}
		return fmt.Errorf("Expected all resources to be gone, but found: %s", strings.Join(resourceNames, ", "))
	}

	return nil
}

func (a *LabeledApp) Rename(_ string, _ string) error { return fmt.Errorf("Not supported") }

func (a *LabeledApp) Meta() (Meta, error) { return Meta{}, nil }

func (a *LabeledApp) Changes() ([]Change, error)             { return nil, nil }
func (a *LabeledApp) LastChange() (Change, error)            { return nil, nil }
func (a *LabeledApp) BeginChange(ChangeMeta) (Change, error) { return NoopChange{}, nil }
func (a *LabeledApp) GCChanges(max int, reviewFunc func(changesToDelete []Change) error) (int, int, error) {
	return 0, 0, nil
}
