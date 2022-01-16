// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	cmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/client-go/kubernetes"
)

type FactorySupportObjs struct {
	CoreClient          kubernetes.Interface
	ResourceTypes       *ctlres.ResourceTypesImpl
	IdentifiedResources ctlres.IdentifiedResources
	Apps                ctlapp.Apps
}

func FactoryClients(depsFactory cmdcore.DepsFactory, nsFlags cmdcore.NamespaceFlags,
	resTypesFlags ResourceTypesFlags, logger logger.Logger) (FactorySupportObjs, error) {

	coreClient, err := depsFactory.CoreClient()
	if err != nil {
		return FactorySupportObjs{}, err
	}

	dynamicClient, err := depsFactory.DynamicClient(cmdcore.DynamicClientOpts{Warnings: true})
	if err != nil {
		return FactorySupportObjs{}, err
	}

	mutedDynamicClient, err := depsFactory.DynamicClient(cmdcore.DynamicClientOpts{Warnings: false})
	if err != nil {
		return FactorySupportObjs{}, err
	}

	resTypes := ctlres.NewResourceTypesImpl(coreClient, ctlres.ResourceTypesImplOpts{
		IgnoreFailingAPIServices:   resTypesFlags.IgnoreFailingAPIServices,
		CanIgnoreFailingAPIService: resTypesFlags.CanIgnoreFailingAPIService,
	})

	resourcesImplOpts := ctlres.ResourcesImplOpts{
		FallbackAllowedNamespaces:        []string{nsFlags.Name},
		ScopeToFallbackAllowedNamespaces: resTypesFlags.ScopeToFallbackAllowedNamespaces,
	}

	resources := ctlres.NewResourcesImpl(
		resTypes, coreClient, dynamicClient, mutedDynamicClient, resourcesImplOpts, logger)

	identifiedResources := ctlres.NewIdentifiedResources(
		coreClient, resTypes, resources, resourcesImplOpts.FallbackAllowedNamespaces, logger)

	result := FactorySupportObjs{
		CoreClient:          coreClient,
		ResourceTypes:       resTypes,
		IdentifiedResources: identifiedResources,
		Apps:                ctlapp.NewApps(nsFlags.Name, coreClient, identifiedResources, logger),
	}

	return result, nil
}

func Factory(depsFactory cmdcore.DepsFactory, appFlags Flags,
	resTypesFlags ResourceTypesFlags, logger logger.Logger) (ctlapp.App, FactorySupportObjs, error) {

	supportingObjs, err := FactoryClients(depsFactory, appFlags.NamespaceFlags, resTypesFlags, logger)
	if err != nil {
		return nil, FactorySupportObjs{}, err
	}

	app, err := supportingObjs.Apps.Find(appFlags.Name)
	if err != nil {
		return nil, FactorySupportObjs{}, err
	}

	return app, supportingObjs, nil
}
