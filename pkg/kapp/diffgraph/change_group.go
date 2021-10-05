// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package diffgraph

import (
	"fmt"
	"strings"

	k8sval "k8s.io/apimachinery/pkg/util/validation"
)

type ChangeGroup struct {
	Name string
}

func MustNewChangeGroupFromAnnString(ann string) ChangeGroup {
	key, err := NewChangeGroupFromAnnString(ann)
	if err != nil {
		panic(err.Error())
	}
	return key
}

func NewChangeGroupFromAnnString(ann string) (ChangeGroup, error) {
	key := ChangeGroup{ann}

	err := key.Validate()
	if err != nil {
		return ChangeGroup{}, err
	}

	return key, nil
}

func (r ChangeGroup) IsEqual(other ChangeGroup) bool {
	return r.Name == other.Name
}

func (r ChangeGroup) Validate() error {
	if len(r.Name) == 0 {
		return fmt.Errorf("Expected non-empty group name")
	}
	errStrs := k8sval.IsQualifiedName(r.Name)
	if len(errStrs) > 0 {
		return fmt.Errorf("Expected change group name %q to be a qualified name: %s", r.Name, strings.Join(errStrs, "; "))
	}
	return nil
}
