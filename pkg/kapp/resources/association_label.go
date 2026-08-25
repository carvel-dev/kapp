// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"crypto/fips140"
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
	//
	// MD5 here is a non-security convenience hash used to keep the label
	// value short and stable; it is not used for authentication or
	// integrity verification. WithoutEnforcement lets this run under
	// GODEBUG=fips140=only, which otherwise panics on any non-approved
	// primitive regardless of how it's used.
	var sum [md5.Size]byte
	fips140.WithoutEnforcement(func() {
		sum = md5.Sum([]byte(NewUniqueResourceKey(a.resource).String()))
	})
	key := fmt.Sprintf("%x", sum)
	return kappAssociationLabelV1 + "." + key
}

func (a AssociationLabel) Key() string   { return kappAssociationLabelKey }
func (a AssociationLabel) Value() string { return a.v1Value() }

func (a AssociationLabel) AsSelector() labels.Selector {
	return labels.Set(map[string]string{kappAssociationLabelKey: a.v1Value()}).AsSelector()
}
