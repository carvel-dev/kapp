package core

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type DepsFactory interface {
	DynamicClient() (dynamic.Interface, error)
	CoreClient() (kubernetes.Interface, error)
}

type DepsFactoryImpl struct {
	configFactory ConfigFactory
}

var _ DepsFactory = &DepsFactoryImpl{}

func NewDepsFactoryImpl(configFactory ConfigFactory) *DepsFactoryImpl {
	return &DepsFactoryImpl{configFactory}
}

func (f *DepsFactoryImpl) DynamicClient() (dynamic.Interface, error) {
	config, err := f.configFactory.RESTConfig()
	if err != nil {
		return nil, err
	}

	// TODO high QPS
	config.QPS = 1000
	config.Burst = 1000

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Building Dynamic clientset: %s", err)
	}

	return clientset, nil
}

func (f *DepsFactoryImpl) CoreClient() (kubernetes.Interface, error) {
	config, err := f.configFactory.RESTConfig()
	if err != nil {
		return nil, err
	}

	config.QPS = 1000
	config.Burst = 1000

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Building Core clientset: %s", err)
	}

	return clientset, nil
}
