package app

import (
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func appFactory(depsFactory cmdcore.DepsFactory, appFlags AppFlags) (
	ctlapp.App, kubernetes.Interface, dynamic.Interface, error) {

	coreClient, err := depsFactory.CoreClient()
	if err != nil {
		return nil, nil, nil, err
	}

	dynamicClient, err := depsFactory.DynamicClient()
	if err != nil {
		return nil, nil, nil, err
	}

	apps := ctlapp.NewApps(appFlags.NamespaceFlags.Name, coreClient, dynamicClient)

	app, err := apps.Find(appFlags.Name)
	if err != nil {
		return nil, nil, nil, err
	}

	return app, coreClient, dynamicClient, nil
}
