// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type ChangeSetFactory struct {
	opts          ChangeSetOpts
	changeFactory ChangeFactory
}

func NewChangeSetFactory(opts ChangeSetOpts, changeFactory ChangeFactory) ChangeSetFactory {
	return ChangeSetFactory{opts, changeFactory}
}

func (f ChangeSetFactory) New(existingRs, newRs []ctlres.Resource) *ChangeSet {
	return NewChangeSet(existingRs, newRs, f.opts, f.changeFactory)
}
