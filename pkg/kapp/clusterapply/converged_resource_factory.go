package clusterapply

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	ctlresm "github.com/k14s/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	associatedRsFunc func(ctlres.Resource, []ctlres.ResourceRef) ([]ctlres.Resource, error)) ConvergedResource {

	specificResFactories := []SpecificResFactory{
		// custom resource waiting behaviour
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewCustomResource(res), nil
		},
		// kapp-controller app resource waiter deals with reconciliation _and_ deletion
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewKappctrlK14sIoV1alpha1App(res), nil
		},
		// Deal with deletion generically since below resource waiters do not not know about that
		// TODO shoud we make all of them deal with deletion internally?
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewDeleting(res), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewApiExtensionsVxCRD(res), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewAPIRegistrationV1APIService(res, f.opts.IgnoreFailingAPIServices), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewAPIRegistrationV1Beta1APIService(res, f.opts.IgnoreFailingAPIServices), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewCoreV1Pod(res), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewCoreV1Service(res), nil
		},
		func(res ctlres.Resource, aRs []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			// Use newly provided associated resources as they may be modified by ConvergedResource
			return ctlresm.NewAppsV1Deployment(res, aRs), []ctlres.ResourceRef{
				{schema.GroupVersionResource{Group: "apps"}}, // for ReplicaSets
				{schema.GroupVersionResource{Group: ""}},     // for Pods
			}
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewAppsV1DaemonSet(res), []ctlres.ResourceRef{
				{schema.GroupVersionResource{Group: "apps"}},
				{schema.GroupVersionResource{Group: ""}}, // for Pods
			}
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewBatchV1Job(res), []ctlres.ResourceRef{
				{schema.GroupVersionResource{Group: "batch"}},
				{schema.GroupVersionResource{Group: ""}}, // for Pods
			}
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewBatchVxCronJob(res), nil
		},
	}

	return NewConvergedResource(res, associatedRsFunc, specificResFactories)
}
