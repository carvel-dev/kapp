package app

import (
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/client-go/kubernetes"
)

func AppFactoryClients(depsFactory cmdcore.DepsFactory, nsFlags cmdcore.NamespaceFlags,
	resTypesFlags ResourceTypesFlags, logger logger.Logger) (
	ctlapp.Apps, kubernetes.Interface, ctlres.IdentifiedResources, error) {

	coreClient, err := depsFactory.CoreClient()
	if err != nil {
		return ctlapp.Apps{}, nil, ctlres.IdentifiedResources{}, err
	}

	dynamicClient, err := depsFactory.DynamicClient()
	if err != nil {
		return ctlapp.Apps{}, nil, ctlres.IdentifiedResources{}, err
	}

	resTypes := ctlres.NewResourceTypesImpl(coreClient, ctlres.ResourceTypesImplOpts{
		IgnoreFailingAPIServices: resTypesFlags.IgnoreFailingAPIServices,
	})

	identifiedResources := ctlres.NewIdentifiedResources(
		coreClient, dynamicClient, resTypes, []string{nsFlags.Name}, logger)

	apps := ctlapp.NewApps(nsFlags.Name, coreClient, identifiedResources, logger)

	return apps, coreClient, identifiedResources, nil
}

func AppFactory(depsFactory cmdcore.DepsFactory, appFlags AppFlags,
	resTypesFlags ResourceTypesFlags, logger logger.Logger) (
	ctlapp.App, kubernetes.Interface, ctlres.IdentifiedResources, error) {

	apps, coreClient, identifiedResources, err := AppFactoryClients(
		depsFactory, appFlags.NamespaceFlags, resTypesFlags, logger)
	if err != nil {
		return nil, nil, ctlres.IdentifiedResources{}, err
	}

	app, err := apps.Find(appFlags.Name)
	if err != nil {
		return nil, nil, ctlres.IdentifiedResources{}, err
	}

	return app, coreClient, identifiedResources, nil
}
