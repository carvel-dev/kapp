// Copyright 2024 The Carvel Authors.
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
	errStrs := r.isQualifiedNameWithoutLen(r.Name)
	if len(errStrs) > 0 {
		return fmt.Errorf("Expected change group name %q to be a qualified name: %s", r.Name, strings.Join(errStrs, "; "))
	}
	return nil
}

func (r ChangeGroup) isQualifiedNameWithoutLen(name string) []string {
	errStrs := k8sval.IsQualifiedName(name)
	var updatedErrStrs []string
	for _, err := range errStrs {
		// Allow change group names to have more characters than the default maxLength
		if !strings.Contains(err, k8sval.MaxLenError(k8sval.DNS1035LabelMaxLength)) {
			updatedErrStrs = append(updatedErrStrs, err)
		}
	}
	return updatedErrStrs
}
