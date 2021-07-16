// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/cppforlife/go-cli-ui/ui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type DepsFactory interface {
	DynamicClient() (dynamic.Interface, error)
	CoreClient() (kubernetes.Interface, error)
}

type DepsFactoryImpl struct {
	configFactory   ConfigFactory
	ui              ui.UI
	printTargetOnce *sync.Once
}

var _ DepsFactory = &DepsFactoryImpl{}

func NewDepsFactoryImpl(configFactory ConfigFactory, ui ui.UI) *DepsFactoryImpl {
	return &DepsFactoryImpl{configFactory, ui, &sync.Once{}}
}

func (f *DepsFactoryImpl) DynamicClient() (dynamic.Interface, error) {
	config, err := f.configFactory.RESTConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Building Dynamic clientset: %s", err)
	}

	f.printTarget(config)

	return clientset, nil
}

func (f *DepsFactoryImpl) CoreClient() (kubernetes.Interface, error) {
	config, err := f.configFactory.RESTConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Building Core clientset: %s", err)
	}

	f.printTarget(config)

	return clientset, nil
}

func (f *DepsFactoryImpl) printTarget(config *rest.Config) {
	f.printTargetOnce.Do(func() {
		nodesDesc := f.summarizeNodes(config)
		if len(nodesDesc) > 0 {
			nodesDesc = fmt.Sprintf(" (nodes: %s)", nodesDesc)
		}
		f.ui.PrintLinef("Target cluster '%s'%s", config.Host, nodesDesc)
	})
}

func (f *DepsFactoryImpl) summarizeNodes(config *rest.Config) string {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ""
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ""
	}

	switch len(nodes.Items) {
	case 0:
		return ""

	case 1:
		return nodes.Items[0].Name

	default:
		oldestNode := nodes.Items[0]
		for _, node := range nodes.Items {
			if node.CreationTimestamp.Before(&oldestNode.CreationTimestamp) {
				oldestNode = node
			}
		}
		return fmt.Sprintf("%s, %d+", oldestNode.Name, len(nodes.Items)-1)
	}
}
