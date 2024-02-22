// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type App interface {
	Name() string
	Namespace() string
	CreationTimestamp() time.Time
	Description() string
	Meta() (Meta, error)

	LabelSelector() (labels.Selector, error)
	UsedGVs() ([]schema.GroupVersion, error)
	UsedGKs() (*[]schema.GroupKind, error)
	UpdateUsedGVsAndGKs([]schema.GroupVersion, []schema.GroupKind) error

	CreateOrUpdate(string, map[string]string, bool) (bool, error)
	Exists() (bool, string, error)
	Delete(bool) error
	Rename(string, string) error

	// Sorted as first is oldest
	Changes() ([]Change, error)
	LastChange() (Change, error)
	BeginChange(ChangeMeta, int) (Change, error)
	GCChanges(max int, reviewFunc func(changesToDelete []Change) error) (int, int, error)
}

type Change interface {
	Name() string
	Meta() ChangeMeta

	Fail() error
	Succeed() error

	Delete() error
}
