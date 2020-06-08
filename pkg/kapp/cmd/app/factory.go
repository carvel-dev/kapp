package app

import (
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/client-go/kubernetes"
)

type AppFactorySupportObjs struct {
	CoreClient          kubernetes.Interface
	ResourceTypes       *ctlres.ResourceTypesImpl
	IdentifiedResources ctlres.IdentifiedResources
	Apps                ctlapp.Apps
}

func AppFactoryClients(depsFactory cmdcore.DepsFactory, nsFlags cmdcore.NamespaceFlags,
	resTypesFlags ResourceTypesFlags, logger logger.Logger) (AppFactorySupportObjs, error) {

	coreClient, err := depsFactory.CoreClient()
	if err != nil {
		return AppFactorySupportObjs{}, err
	}

	dynamicClient, err := depsFactory.DynamicClient()
	if err != nil {
		return AppFactorySupportObjs{}, err
	}

	fallbackAllowedNss := []string{nsFlags.Name}

	resTypes := ctlres.NewResourceTypesImpl(coreClient, ctlres.ResourceTypesImplOpts{
		IgnoreFailingAPIServices:   resTypesFlags.IgnoreFailingAPIServices,
		CanIgnoreFailingAPIService: resTypesFlags.CanIgnoreFailingAPIService,
	})

	resources := ctlres.NewResources(resTypes, coreClient, dynamicClient, fallbackAllowedNss, logger)

	identifiedResources := ctlres.NewIdentifiedResources(
		coreClient, resTypes, resources, fallbackAllowedNss, logger)

	result := AppFactorySupportObjs{
		CoreClient:          coreClient,
		ResourceTypes:       resTypes,
		IdentifiedResources: identifiedResources,
		Apps:                ctlapp.NewApps(nsFlags.Name, coreClient, identifiedResources, logger),
	}

	return result, nil
}

func AppFactory(depsFactory cmdcore.DepsFactory, appFlags AppFlags,
	resTypesFlags ResourceTypesFlags, logger logger.Logger) (ctlapp.App, AppFactorySupportObjs, error) {

	supportingObjs, err := AppFactoryClients(depsFactory, appFlags.NamespaceFlags, resTypesFlags, logger)
	if err != nil {
		return nil, AppFactorySupportObjs{}, err
	}

	app, err := supportingObjs.Apps.Find(appFlags.Name)
	if err != nil {
		return nil, AppFactorySupportObjs{}, err
	}

	return app, supportingObjs, nil
}
