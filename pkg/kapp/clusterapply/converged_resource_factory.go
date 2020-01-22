package clusterapply

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
)

type ConvergedResourceFactoryOpts struct {
	IgnoreFailingAPIServices bool
}

type ConvergedResourceFactory struct {
	opts ConvergedResourceFactoryOpts
}

func NewConvergedResourceFactory(
	opts ConvergedResourceFactoryOpts) ConvergedResourceFactory {
	return ConvergedResourceFactory{opts}
}

func (f ConvergedResourceFactory) New(res ctlres.Resource,
	associatedRs []ctlres.Resource) ConvergedResource {

	specificResFactories := []SpecificResFactory{
		// kapp-controller app resource waiter deals with reconciliation _and_ deletion
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource {
			return ctlresm.NewKappctrlK14sIoV1alpha1App(res)
		},

		// Deal with deletion generically since below resource waiters do not not know about that
		// TODO shoud we make all of them deal with deletion internally?
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource { return ctlresm.NewDeleting(res) },

		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource {
			return ctlresm.NewApiExtensionsVxCRD(res)
		},
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource {
			return ctlresm.NewAPIRegistrationV1APIService(res, f.opts.IgnoreFailingAPIServices)
		},
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource {
			return ctlresm.NewAPIRegistrationV1Beta1APIService(res, f.opts.IgnoreFailingAPIServices)
		},
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource { return ctlresm.NewCoreV1Pod(res) },
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource { return ctlresm.NewCoreV1Service(res) },
		func(res ctlres.Resource, aRs []ctlres.Resource) SpecificResource {
			// Use newly provided associated resources as they may be modified by ConvergedResource
			return ctlresm.NewAppsV1Deployment(res, aRs)
		},
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource {
			return ctlresm.NewAppsV1DaemonSet(res)
		},
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource { return ctlresm.NewBatchV1Job(res) },
		func(res ctlres.Resource, _ []ctlres.Resource) SpecificResource { return ctlresm.NewBatchVxCronJob(res) },
	}

	return NewConvergedResource(res, associatedRs, specificResFactories)
}
