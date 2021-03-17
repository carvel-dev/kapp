// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"crypto/md5"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

const (
	kappAssociationLabelKey = "kapp.k14s.io/association"
	kappAssociationLabelV1  = "v1"
)

type AssociationLabel struct {
	resource Resource
}

func NewAssociationLabel(resource Resource) AssociationLabel {
	return AssociationLabel{resource}
}

func (a AssociationLabel) v1Value() string {
	// max 63 char for label values
	key := fmt.Sprintf("%x", md5.Sum([]byte(NewUniqueResourceKey(a.resource).String())))
	return kappAssociationLabelV1 + "." + key
}

func (a AssociationLabel) Key() string   { return kappAssociationLabelKey }
func (a AssociationLabel) Value() string { return a.v1Value() }

func (a AssociationLabel) AsSelector() labels.Selector {
	return labels.Set(map[string]string{kappAssociationLabelKey: a.v1Value()}).AsSelector()
}
