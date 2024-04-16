// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package clusterapply

import (
	ctlconf "carvel.dev/kapp/pkg/kapp/config"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	ctlresm "carvel.dev/kapp/pkg/kapp/resourcesmisc"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ConvergedResourceFactoryOpts struct {
	IgnoreFailingAPIServices bool
}

type ConvergedResourceFactory struct {
	waitRules []ctlconf.WaitRule
	opts      ConvergedResourceFactoryOpts
}

func NewConvergedResourceFactory(waitRules []ctlconf.WaitRule,
	opts ConvergedResourceFactoryOpts) ConvergedResourceFactory {
	return ConvergedResourceFactory{waitRules, opts}
}

func (f ConvergedResourceFactory) New(res ctlres.Resource,
	associatedRsFunc func(ctlres.Resource, []ctlres.ResourceRef) ([]ctlres.Resource, error)) ConvergedResource {

	specificResFactories := []SpecificResFactory{
		// kapp-controller app resource waiter deals with reconciliation _and_ deletion
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewKappctrlK14sIoV1alpha1App(res), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewPackagingCarvelDevV1alpha1PackageInstall(res), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewPackagingCarvelDevV1alpha1PackageRepo(res), nil
		},
		// Deal with deletion generically since below resource waiters do not not know about that
		// TODO shoud we make all of them deal with deletion internally?
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewDeleting(res), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewCustomWaitingResource(res, f.waitRules), nil
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewAPIExtensionsVxCRD(res), nil
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
				{schema.GroupVersionResource{Group: "apps", Resource: "replicasets"}},
				{schema.GroupVersionResource{Group: "", Resource: "pods"}},
			}
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewAppsV1DaemonSet(res), []ctlres.ResourceRef{
				{schema.GroupVersionResource{Group: "apps", Resource: "replicasets"}},
				{schema.GroupVersionResource{Group: "", Resource: "pods"}},
			}
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewBatchV1Job(res), []ctlres.ResourceRef{
				{schema.GroupVersionResource{Group: "batch", Resource: "jobs"}},
				{schema.GroupVersionResource{Group: "", Resource: "pods"}},
			}
		},
		func(res ctlres.Resource, _ []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewBatchVxCronJob(res), nil
		},
		func(res ctlres.Resource, aRs []ctlres.Resource) (SpecificResource, []ctlres.ResourceRef) {
			return ctlresm.NewAppsV1StatefulSet(res, aRs), []ctlres.ResourceRef{
				{schema.GroupVersionResource{Group: "", Resource: "persistentvolumeclaims"}},
				{schema.GroupVersionResource{Group: "", Resource: "pods"}},
				// omit ControllerRevisions: we'll rarely (if ever) wait on them; reporting on them is noise
			}
		},
	}

	return NewConvergedResource(res, associatedRsFunc, specificResFactories)
}
