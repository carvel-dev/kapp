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
	associatedRsFunc func(ctlres.Resource) ([]ctlres.Resource, error)) ConvergedResource {

	specificResFactories := []SpecificResFactory{
		// kapp-controller app resource waiter deals with reconciliation _and_ deletion
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewKappctrlK14sIoV1alpha1App(res), false
		},
		// Deal with deletion generically since below resource waiters do not not know about that
		// TODO shoud we make all of them deal with deletion internally?
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewDeleting(res), false
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewApiExtensionsVxCRD(res), false
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewAPIRegistrationV1APIService(res, f.opts.IgnoreFailingAPIServices), false
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewAPIRegistrationV1Beta1APIService(res, f.opts.IgnoreFailingAPIServices), false
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewCoreV1Pod(res), false
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewCoreV1Service(res), false
		},
		func(res ctlres.Resource, aRs []ctlres.Resource) (SpecificResource, bool) {
			// Use newly provided associated resources as they may be modified by ConvergedResource
			return ctlresm.NewAppsV1Deployment(res, aRs), true
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewAppsV1DaemonSet(res), true
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewBatchV1Job(res), true
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, bool) {
			return ctlresm.NewBatchVxCronJob(res), false
		},
	}

	return NewConvergedResource(res, associatedRsFunc, specificResFactories)
}
