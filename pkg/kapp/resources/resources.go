// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/k14s/kapp/pkg/kapp/util"
	"golang.org/x/net/http2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// type ResourceInterface interface {
// 	Create(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)
// 	Update(obj *unstructured.Unstructured, subresources ...string) (*unstructured.Unstructured, error)
// 	UpdateStatus(obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
// 	Delete(name string, options *metav1.DeleteOptions, subresources ...string) error
// 	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
// 	Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
// 	List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error)
// 	Watch(opts metav1.ListOptions) (watch.Interface, error)
// 	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*unstructured.Unstructured, error)
// }

const (
	resourcesDebug = false
)

type Resources interface {
	All([]ResourceType, AllOpts) ([]Resource, error)
	Delete(Resource) error
	Exists(Resource, ExistsOpts) (Resource, bool, error)
	Get(Resource) (Resource, error)
	Patch(Resource, types.PatchType, []byte) (Resource, error)
	Update(Resource) (Resource, error)
	Create(resource Resource) (Resource, error)
}

type ExistsOpts struct {
	SameUID bool
}

type ResourcesImpl struct {
	resourceTypes      ResourceTypes
	coreClient         kubernetes.Interface
	dynamicClient      dynamic.Interface
	mutedDynamicClient dynamic.Interface
	opts               ResourcesImplOpts

	assumedAllowedNamespacesMemoLock sync.Mutex
	assumedAllowedNamespacesMemo     *[]string

	logger logger.Logger
}

type ResourcesImplOpts struct {
	FallbackAllowedNamespaces        []string
	ScopeToFallbackAllowedNamespaces bool
}

func NewResourcesImpl(resourceTypes ResourceTypes, coreClient kubernetes.Interface,
	dynamicClient dynamic.Interface, mutedDynamicClient dynamic.Interface,
	opts ResourcesImplOpts, logger logger.Logger) *ResourcesImpl {

	return &ResourcesImpl{
		resourceTypes:      resourceTypes,
		coreClient:         coreClient,
		dynamicClient:      dynamicClient,
		mutedDynamicClient: mutedDynamicClient,
		opts:               opts,
		logger:             logger.NewPrefixed("Resources"),
	}
}

type unstructItems struct {
	ResType ResourceType
	Items   []unstructured.Unstructured
}

func (c *ResourcesImpl) All(resTypes []ResourceType, opts AllOpts) ([]Resource, error) {
	defer c.logger.DebugFunc("All").Finish()

	if opts.ListOpts == nil {
		opts.ListOpts = &metav1.ListOptions{}
	}

	nsScope := "" // all namespaces by default
	nsScopeLimited := c.opts.ScopeToFallbackAllowedNamespaces && len(c.opts.FallbackAllowedNamespaces) == 1

	// Eagerly use single fallback namespace to avoid making all-namespaces request
	// just to see it fail, and fallback to making namespace-scoped request
	if nsScopeLimited {
		nsScope = c.opts.FallbackAllowedNamespaces[0]
		c.logger.Info("Scoping listings to single namespace: %s", nsScope)
	}

	unstructItemsCh := make(chan unstructItems, len(resTypes))
	fatalErrsCh := make(chan error, len(resTypes))
	var itemsDone sync.WaitGroup

	for _, resType := range resTypes {
		resType := resType // copy
		itemsDone.Add(1)

		go func() {
			defer itemsDone.Done()

			defer c.logger.DebugFunc(resType.GroupVersionResource.String()).Finish()

			var list *unstructured.UnstructuredList
			var err error

			client := c.mutedDynamicClient.Resource(resType.GroupVersionResource)

			err = util.Retry2(time.Second, 5*time.Second, c.isServerRescaleErr, func() error {
				if resType.Namespaced() {
					list, err = client.Namespace(nsScope).List(context.TODO(), *opts.ListOpts)
				} else {
					list, err = client.List(context.TODO(), *opts.ListOpts)
				}
				return err
			})

			if err != nil {
				if !errors.IsForbidden(err) {
					// Ignore certain GVs due to failing API backing
					if c.resourceTypes.CanIgnoreFailingGroupVersion(resType.GroupVersion()) {
						c.logger.Info("Ignoring group version: %#v: %s", resType.GroupVersionResource, err)
					} else {
						fatalErrsCh <- fmt.Errorf("Listing %#v, namespaced: %t: %s", resType.GroupVersionResource, resType.Namespaced(), err)
					}
					return
				}
				// At this point err==Forbidden...

				// In case ns scope is limited already, we will not gain anything
				// by trying to run namespace scoped lists for allowed namespaced
				// (ie since it's would be same request that just failed)
				if !resType.Namespaced() || nsScopeLimited {
					c.logger.Debug("Skipping forbidden group version: %#v", resType.GroupVersionResource)
					return
				}

				// TODO improve perf somehow
				list, err = c.allForNamespaces(client, opts.ListOpts)
				if err != nil {
					// Ignore certain GVs due to failing API backing
					if c.resourceTypes.CanIgnoreFailingGroupVersion(resType.GroupVersion()) {
						c.logger.Info("Ignoring group version: %#v", resType.GroupVersionResource)
					} else {
						fatalErrsCh <- fmt.Errorf("Listing %#v, namespaced: %t: %s", resType.GroupVersionResource, resType.Namespaced(), err)
					}
					return
				}
			}

			unstructItemsCh <- unstructItems{resType, list.Items}
		}()
	}

	itemsDone.Wait()
	close(unstructItemsCh)
	close(fatalErrsCh)

	for err := range fatalErrsCh {
		return nil, err // TODO consolidate
	}

	var resources []Resource

	for unstructItem := range unstructItemsCh {
		for _, item := range unstructItem.Items {
			resources = append(resources, NewResourceUnstructured(item, unstructItem.ResType))
		}
	}

	return resources, nil
}

func (c *ResourcesImpl) allForNamespaces(client dynamic.NamespaceableResourceInterface, listOpts *metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	defer c.logger.DebugFunc("allForNamespaces").Finish()

	allowedNs, err := c.assumedAllowedNamespaces()
	if err != nil {
		return nil, err
	}

	var itemsDone sync.WaitGroup
	fatalErrsCh := make(chan error, len(allowedNs))
	unstructItemsCh := make(chan *unstructured.UnstructuredList, len(allowedNs))

	for _, ns := range allowedNs {
		ns := ns // copy
		itemsDone.Add(1)

		go func() {
			defer itemsDone.Done()

			resList, err := client.Namespace(ns).List(context.TODO(), *listOpts)
			if err != nil {
				if !errors.IsForbidden(err) {
					fatalErrsCh <- err
					return
				}
				// Ignore forbidden errors
				// TODO somehow surface them
			} else {
				unstructItemsCh <- resList
			}
		}()
	}

	itemsDone.Wait()
	close(fatalErrsCh)
	close(unstructItemsCh)

	for fatalErr := range fatalErrsCh {
		return nil, fatalErr
	}

	list := &unstructured.UnstructuredList{}

	for resList := range unstructItemsCh {
		list.Items = append(list.Items, resList.Items...)
	}

	return list, nil
}

func (c *ResourcesImpl) Create(resource Resource) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { c.logger.Debug("create %s", time.Now().UTC().Sub(t1)) }()

		bs, _ := resource.AsYAMLBytes()
		c.logger.Debug("create resource %s\n%s\n", resource.Description(), bs)
	}

	resClient, resType, err := c.resourceClient(resource, resourceClientOpts{Warnings: true})
	if err != nil {
		return nil, err
	}

	var createdUn *unstructured.Unstructured

	err = util.Retry2(time.Second, 5*time.Second, c.isGeneralRetryableErr, func() error {
		createdUn, err = resClient.Create(context.TODO(), resource.unstructuredPtr(), metav1.CreateOptions{})
		return err
	})
	if err != nil {
		return nil, c.resourceErr(err, "Creating", resource)
	}

	return NewResourceUnstructured(*createdUn, resType), nil
}

func (c *ResourcesImpl) Update(resource Resource) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { c.logger.Debug("update %s", time.Now().UTC().Sub(t1)) }()

		bs, _ := resource.AsYAMLBytes()
		c.logger.Debug("update resource %s\n%s\n", resource.Description(), bs)
	}

	resClient, resType, err := c.resourceClient(resource, resourceClientOpts{Warnings: true})
	if err != nil {
		return nil, err
	}

	var updatedUn *unstructured.Unstructured

	err = util.Retry2(time.Second, 5*time.Second, c.isGeneralRetryableErr, func() error {
		updatedUn, err = resClient.Update(context.TODO(), resource.unstructuredPtr(), metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return nil, c.resourceErr(err, "Updating", resource)
	}

	return NewResourceUnstructured(*updatedUn, resType), nil
}

func (c *ResourcesImpl) Patch(resource Resource, patchType types.PatchType, data []byte) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { c.logger.Debug("patch %s", time.Now().UTC().Sub(t1)) }()
	}

	resClient, resType, err := c.resourceClient(resource, resourceClientOpts{Warnings: true})
	if err != nil {
		return nil, err
	}

	var patchedUn *unstructured.Unstructured

	err = util.Retry2(time.Second, 5*time.Second, c.isGeneralRetryableErr, func() error {
		patchedUn, err = resClient.Patch(context.TODO(), resource.Name(), patchType, data, metav1.PatchOptions{})
		return err
	})
	if err != nil {
		return nil, c.resourceErr(err, "Patching", resource)
	}

	return NewResourceUnstructured(*patchedUn, resType), nil
}

func (c *ResourcesImpl) Delete(resource Resource) error {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { c.logger.Debug("delete %s", time.Now().UTC().Sub(t1)) }()
	}

	if resource.IsDeleting() {
		c.logger.Info("TODO resource '%s' is already deleting", resource.Description())
		return nil
	}

	resClient, resType, err := c.resourceClient(resource, resourceClientOpts{Warnings: true})
	if err != nil {
		return err
	}

	if resType.Deletable() {
		// TODO is setting deletion policy a correct thing to do?
		// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#setting-the-cascading-deletion-policy
		delPol := metav1.DeletePropagationBackground
		delOpts := metav1.DeleteOptions{PropagationPolicy: &delPol}

		// Some resources may not have UID (example: PodMetrics.metrics.k8s.io)
		resUID := types.UID(resource.UID())
		if len(resUID) > 0 {
			delOpts.Preconditions = &metav1.Preconditions{UID: &resUID}
		}

		err = resClient.Delete(context.TODO(), resource.Name(), delOpts)
		if err != nil {
			if errors.IsNotFound(err) {
				c.logger.Info("TODO resource '%s' is already gone", resource.Description())
				return nil
			}
			if c.isPodMetrics(resource, err) {
				return nil
			}
			return c.resourceErr(err, "Deleting", resource)
		}
	} else {
		c.logger.Info("TODO resource '%s' is not deletable", resource.Description()) // TODO
	}

	return nil
}

func (c *ResourcesImpl) Get(resource Resource) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { c.logger.Debug("get %s", time.Now().UTC().Sub(t1)) }()
	}

	resClient, resType, err := c.resourceClient(resource, resourceClientOpts{Warnings: false})
	if err != nil {
		return nil, err
	}

	var item *unstructured.Unstructured

	err = util.Retry2(time.Second, 5*time.Second, c.isServerRescaleErr, func() error {
		var err error
		item, err = resClient.Get(context.TODO(), resource.Name(), metav1.GetOptions{})
		return err
	})
	if err != nil {
		return nil, c.resourceErr(err, "Getting", resource)
	}

	return NewResourceUnstructured(*item, resType), nil
}

func (c *ResourcesImpl) Exists(resource Resource, existsOpts ExistsOpts) (Resource, bool, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { c.logger.Debug("exists %s", time.Now().UTC().Sub(t1)) }()
	}

	resClient, resType, err := c.resourceClient(resource, resourceClientOpts{Warnings: false})
	if err != nil {
		// Assume if type is not known to the API server
		// then such resource cannot exist on the server
		if _, ok := err.(ResourceTypesUnknownTypeErr); ok {
			return nil, false, nil
		}
		return nil, false, err
	}

	var found bool
	var resObj Resource

	err = util.Retry(time.Second, time.Minute, func() (bool, error) {
		fetchedRes, err := resClient.Get(context.TODO(), resource.Name(), metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				found = false
				return true, nil
			}
			if c.isPodMetrics(resource, err) {
				found = false
				return true, nil
			}
			if c.isServerRescaleErr(err) {
				return false, nil
			}
			// No point in waiting if we are not allowed to get it
			isDone := errors.IsForbidden(err)
			// TODO sometimes metav1.StatusReasonUnknown is returned (empty string)
			// might be related to deletion of mutating webhook
			return isDone, c.resourceErr(err, "Checking existence of", resource)
		}

		// Check if we have to compare the UID's also to confirm if it is same resource.
		if existsOpts.SameUID {
			if fetchedRes != nil {
				if string(fetchedRes.GetUID()) != resource.UID() {
					found = false
					return true, nil
				}
			}
		}

		found = true
		resObj = NewResourceUnstructured(*fetchedRes, resType)
		return true, nil
	})

	return resObj, found, err
}

var (
	// Error example: Checking existence of resource podmetrics/knative-ingressgateway-646d475cbb-c82qb (metrics.k8s.io/v1beta1)
	//   namespace: istio-system: Error while getting pod knative-ingressgateway-646d475cbb-c82qb:
	//   pod "knative-ingressgateway-646d475cbb-c82qb" not found (reason: )
	// Note that it says pod is not found even though we were checking on podmetrics.
	// (https://github.com/kubernetes-sigs/metrics-server/blob/8d7aca3c6d770bc37d93515bf731a08332b8025b/pkg/api/pod.go#L133)
	podMetricsNotFoundErrCheck = regexp.MustCompile("Error while getting pod (.+) not found \\(reason: \\)")
)

func (c *ResourcesImpl) isPodMetrics(resource Resource, err error) bool {
	// Abnormal error case. Get/Delete on PodMetrics may fail
	// without NotFound reason due to its dependence on Pod existence
	if resource.Kind() == "PodMetrics" && resource.APIGroup() == "metrics.k8s.io" {
		if podMetricsNotFoundErrCheck.MatchString(err.Error()) {
			return true
		}
	}
	return false
}

func (c *ResourcesImpl) isGeneralRetryableErr(err error) bool {
	return IsResourceChangeBlockedErr(err) || c.isServerRescaleErr(err) || c.isEtcdRetryableError(err) ||
		c.isResourceQuotaConflict(err) || c.isInternalFailure(err) || errors.IsTooManyRequests(err)
}

// Fixes issues I observed with GKE:
// Operation cannot be fulfilled on resourcequotas "gke-resource-quotas": the object has been modified;
// please apply your changes to the latest version and try again (reason: Conflict)
// Works around: https://github.com/kubernetes/kubernetes/issues/67761 by retrying.
func (c *ResourcesImpl) isResourceQuotaConflict(err error) bool {
	return errors.IsConflict(err) && strings.Contains(err.Error(), "Operation cannot be fulfilled on resourcequota")
}

func (c *ResourcesImpl) isServerRescaleErr(err error) bool {
	switch err := err.(type) {
	case *http2.GoAwayError:
		return true
	case *errors.StatusError:
		if err.ErrStatus.Reason == metav1.StatusReasonServiceUnavailable {
			return true
		}
	}
	return false
}

// Handles case pointed out in : https://github.com/vmware-tanzu/carvel-kapp/issues/258.
// An internal network error which might succeed on retrying.
func (c *ResourcesImpl) isInternalFailure(err error) bool {
	switch err := err.(type) {
	case *errors.StatusError:
		if errors.IsInternalError(err) {
			return true
		}
	}
	return false
}

func (c *ResourcesImpl) resourceErr(err error, action string, resource Resource) error {
	if typedErr, ok := err.(errors.APIStatus); ok {
		return resourceStatusErr{resourcePlainErr{err, action, resource}, typedErr.Status()}
	}
	return resourcePlainErr{err, action, resource}
}

type resourceClientOpts struct {
	Warnings bool
}

func (c *ResourcesImpl) resourceClient(resource Resource, opts resourceClientOpts) (dynamic.ResourceInterface, ResourceType, error) {
	resType, err := c.resourceTypes.Find(resource)
	if err != nil {
		return nil, ResourceType{}, err
	}

	var dynamicClient dynamic.Interface
	if opts.Warnings {
		dynamicClient = c.dynamicClient
	} else {
		dynamicClient = c.mutedDynamicClient
	}

	return dynamicClient.Resource(resType.GroupVersionResource).Namespace(resource.Namespace()), resType, nil
}

func (c *ResourcesImpl) assumedAllowedNamespaces() ([]string, error) {
	c.assumedAllowedNamespacesMemoLock.Lock()
	defer c.assumedAllowedNamespacesMemoLock.Unlock()

	if c.assumedAllowedNamespacesMemo != nil {
		return *c.assumedAllowedNamespacesMemo, nil
	}

	nsList, err := c.coreClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if errors.IsForbidden(err) {
			if len(c.opts.FallbackAllowedNamespaces) > 0 {
				return c.opts.FallbackAllowedNamespaces, nil
			}
		}
		return nil, fmt.Errorf("Fetching all namespaces: %s", err)
	}

	c.logger.Info("Falling back to checking each namespace separately (much slower)")

	var nsNames []string

	for _, ns := range nsList.Items {
		nsNames = append(nsNames, ns.Name)
	}

	c.assumedAllowedNamespacesMemo = &nsNames

	return nsNames, nil
}

var (
	// Error example: conversion webhook for cert-manager.io/v1alpha3, Kind=Issuer failed:
	//   Post https://cert-manager-webhook.cert-manager.svc:443/convert?timeout=30s:
	//   x509: certificate signed by unknown authority (reason: )
	conversionWebhookErrCheck = regexp.MustCompile("conversion webhook for (.+) failed:")

	// Matches retryable etcdserver errors
	// Comprehensive list of errors at : https://github.com/etcd-io/etcd/blob/main/server/etcdserver/errors.go
	etcdserverRetryableErrCheck = regexp.MustCompile("etcdserver:(.+)(leader changed|timed out)")
)

func IsResourceChangeBlockedErr(err error) bool {
	// TODO is there a better way to detect these errors?
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "Internal error occurred: failed calling admission webhook"):
		return true
	case strings.Contains(errMsg, "Internal error occurred: failed calling webhook"):
		return true
	case conversionWebhookErrCheck.MatchString(errMsg):
		return true
	default:
		return false
	}
}

// Retries retryable errors thrown by etcd server.
// Addresses : https://github.com/vmware-tanzu/carvel-kapp/issues/106
func (c *ResourcesImpl) isEtcdRetryableError(err error) bool {
	return etcdserverRetryableErrCheck.MatchString(err.Error())
}

type AllOpts struct {
	ListOpts *metav1.ListOptions
}

type resourceStatusErr struct {
	err    resourcePlainErr
	status metav1.Status
}

var _ errors.APIStatus = resourceStatusErr{}

func (e resourceStatusErr) Error() string         { return e.err.Error() }
func (e resourceStatusErr) Status() metav1.Status { return e.status }

type resourcePlainErr struct {
	err      error
	action   string
	resource Resource
}

func (e resourcePlainErr) Error() string {
	return fmt.Sprintf("%s resource %s: API server says: %s (reason: %s)",
		e.action, e.resource.Description(), e.err, errors.ReasonForError(e.err))
}
