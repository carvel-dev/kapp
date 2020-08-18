// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"strings"
)

const (
	// Version annotation is used to indicate deployed resource version,
	// but is also used to determine transientiveness (kapp deploy vs some controller)
	kappIdentityAnnKey = "kapp.k14s.io/identity"
	kappIdentityAnnV1  = "v1"
)

type IdentityAnnotation struct {
	resource Resource
}

func NewIdentityAnnotation(resource Resource) IdentityAnnotation {
	return IdentityAnnotation{resource}
}

// Valid returns true if signature matches resource itself
func (a IdentityAnnotation) Valid() bool {
	pieces := strings.Split(a.resource.Annotations()[kappIdentityAnnKey], ";")

	switch pieces[0] {
	case kappIdentityAnnV1:
		if len(pieces) != 3 {
			return false
		}
		return NewUniqueResourceKey(a.resource).String() == pieces[1]

	default:
		return false
	}
}

// MatchesVersion returns true if annotation is valid and it matches version
func (a IdentityAnnotation) MatchesVersion() bool {
	if !a.Valid() {
		return false
	}

	pieces := strings.Split(a.resource.Annotations()[kappIdentityAnnKey], ";")

	switch pieces[0] {
	case kappIdentityAnnV1:
		return a.resource.APIVersion() == pieces[2]

	default:
		return false
	}
}

func (a IdentityAnnotation) v1Value() string {
	return kappIdentityAnnV1 + ";" + NewUniqueResourceKey(a.resource).String() + ";" + a.resource.APIVersion()
}

func (a IdentityAnnotation) AddMod() StringMapAppendMod {
	return StringMapAppendMod{
		ResourceMatcher: AllMatcher{},
		Path:            NewPathFromStrings([]string{"metadata", "annotations"}),
		KVs:             map[string]string{kappIdentityAnnKey: a.v1Value()},
	}
}

func (a IdentityAnnotation) RemoveMod() FieldRemoveMod {
	return FieldRemoveMod{
		ResourceMatcher: AllMatcher{},
		Path:            NewPathFromStrings([]string{"metadata", "annotations", kappIdentityAnnKey}),
	}
}
