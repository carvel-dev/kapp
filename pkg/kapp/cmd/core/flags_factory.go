// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package core

type FlagsFactory struct {
	configFactory ConfigFactory
	depsFactory   DepsFactory
}

func NewFlagsFactory(configFactory ConfigFactory, depsFactory DepsFactory) FlagsFactory {
	return FlagsFactory{configFactory, depsFactory}
}

func (f FlagsFactory) NewNamespaceNameFlag(str *string) *NamespaceNameFlag {
	return NewNamespaceNameFlag(str, f.configFactory)
}
