package app

import (
	"k8s.io/apimachinery/pkg/labels"
)

type App interface {
	Name() string
	Namespace() string
	Meta() (AppMeta, error)
	LabelSelector() (labels.Selector, error)

	CreateOrUpdate(map[string]string) error
	Exists() (bool, error)
	Delete() error

	Changes() ([]Change, error)
	LastChange() (Change, error)
	BeginChange(ChangeMeta) (Change, error)
}

type Change interface {
	Name() string
	Meta() ChangeMeta

	Fail() error
	Succeed() error
}
