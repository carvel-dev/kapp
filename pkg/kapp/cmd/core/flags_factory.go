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
