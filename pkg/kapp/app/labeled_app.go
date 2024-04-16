// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"strings"
	"time"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
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

func (a *LabeledApp) CreationTimestamp() time.Time { return time.Time{} }

func (a *LabeledApp) Description() string {
	return fmt.Sprintf("labeled app '%s'", a.Name())
}

func (a *LabeledApp) LabelSelector() (labels.Selector, error) {
	return a.labelSelector, nil
}

func (a *LabeledApp) UsedGVs() ([]schema.GroupVersion, error)                             { return nil, nil }
func (a *LabeledApp) UsedGKs() (*[]schema.GroupKind, error)                               { return nil, nil }
func (a *LabeledApp) UpdateUsedGVsAndGKs([]schema.GroupVersion, []schema.GroupKind) error { return nil }

func (a *LabeledApp) CreateOrUpdate(_ string, _ map[string]string, _ bool) (bool, error) {
	return false, nil
}
func (a *LabeledApp) Exists() (bool, string, error) { return true, "", nil }

func (a *LabeledApp) Delete() error {
	labelSelector, err := a.LabelSelector()
	if err != nil {
		return err
	}

	rs, err := a.identifiedResources.List(labelSelector, nil, ctlres.IdentifiedResourcesListOpts{IgnoreCachedResTypes: true})
	if err != nil {
		return fmt.Errorf("Relisting app resources: %w", err)
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

func (a *LabeledApp) Changes() ([]Change, error)                  { return nil, nil }
func (a *LabeledApp) LastChange() (Change, error)                 { return nil, nil }
func (a *LabeledApp) BeginChange(ChangeMeta, int) (Change, error) { return NoopChange{}, nil }
func (a *LabeledApp) GCChanges(_ int, _ func(changesToDelete []Change) error) (int, int, error) {
	return 0, 0, nil
}
