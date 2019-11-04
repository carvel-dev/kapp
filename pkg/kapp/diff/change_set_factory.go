package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
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
