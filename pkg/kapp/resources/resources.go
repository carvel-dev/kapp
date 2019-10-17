package resources

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/k14s/kapp/pkg/kapp/util"
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

type Resources struct {
	resourceTypes ResourceTypes
	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface

	assumedAllowedNamespacesMemoLock sync.Mutex
	assumedAllowedNamespacesMemo     *[]string

	fallbackAllowedNamespaces []string

	logger logger.Logger
}

func NewResources(resourceTypes ResourceTypes, coreClient kubernetes.Interface,
	dynamicClient dynamic.Interface, fallbackAllowedNamespaces []string, logger logger.Logger) *Resources {

	return &Resources{
		resourceTypes:             resourceTypes,
		coreClient:                coreClient,
		dynamicClient:             dynamicClient,
		fallbackAllowedNamespaces: fallbackAllowedNamespaces,
		logger:                    logger.NewPrefixed("Resources"),
	}
}

type unstructItems struct {
	ResType ResourceType
	Items   []unstructured.Unstructured
}

func (c *Resources) All(resTypes []ResourceType, opts ResourcesAllOpts) ([]Resource, error) {
	defer c.logger.DebugFunc("All").Finish()

	if opts.ListOpts == nil {
		opts.ListOpts = &metav1.ListOptions{}
	}

	unstructItemsCh := make(chan unstructItems, len(resTypes))
	warnErrsCh := make(chan error, len(resTypes))
	fatalErrsCh := make(chan error, len(resTypes))
	var itemsDone sync.WaitGroup

	listThrottle := util.NewThrottle(10)

	for _, resType := range resTypes {
		resType := resType // copy
		itemsDone.Add(1)

		go func() {
			defer itemsDone.Done()

			listThrottle.Take()
			defer listThrottle.Done()

			defer c.logger.DebugFunc(resType.GroupVersionResource.String()).Finish()

			var list *unstructured.UnstructuredList
			var err error

			client := c.dynamicClient.Resource(resType.GroupVersionResource)

			if resType.Namespaced() {
				list, err = client.Namespace("").List(*opts.ListOpts)
			} else {
				list, err = client.List(*opts.ListOpts)
			}

			if err != nil {
				if !errors.IsForbidden(err) {
					fatalErrsCh <- fmt.Errorf("Listing %#v, namespaced: %t: %s", resType.GroupVersionResource, resType.Namespaced(), err)
					return
				}

				if !resType.Namespaced() {
					warnErrsCh <- fmt.Errorf("Listing %#v, namespaced: %t: %s", resType.GroupVersionResource, resType.Namespaced(), err)
					return
				}

				// TODO improve perf somehow
				list, err = c.allForNamespaces(client, opts.ListOpts)
				if err != nil {
					fatalErrsCh <- fmt.Errorf("Listing %#v, namespaced: %t: %s", resType.GroupVersionResource, resType.Namespaced(), err)
					return
				}
			}

			unstructItemsCh <- unstructItems{resType, list.Items}
		}()
	}

	itemsDone.Wait()
	close(unstructItemsCh)
	close(warnErrsCh)
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

func (c *Resources) allForNamespaces(client dynamic.NamespaceableResourceInterface, listOpts *metav1.ListOptions) (*unstructured.UnstructuredList, error) {
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

			resList, err := client.Namespace(ns).List(*listOpts)
			if err != nil {
				if !errors.IsForbidden(err) {
					fatalErrsCh <- err
					return
				} else {
					// TODO warnErrsCh <- fmt.Errorf("Listing %#v, namespaced: %t: %s", resType.GroupVersionResource, resType.Namespaced(), err)
					// Continue trying other namespaces
				}
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

func (c *Resources) Create(resource Resource) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { fmt.Printf("create %s\n", time.Now().UTC().Sub(t1)) }()

		bs, _ := resource.AsYAMLBytes()
		fmt.Printf("create resource %s\n%s\n", resource.Description(), bs)
	}

	resClient, resType, err := c.resourceClient(resource)
	if err != nil {
		return nil, err
	}

	var createdUn *unstructured.Unstructured

	err = util.Retry(time.Second, time.Minute, func() (bool, error) {
		createdUn, err = resClient.Create(resource.unstructuredPtr())
		if err != nil {
			return c.doneRetryingErr(err), c.resourceErr(err, "Creating", resource)
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return NewResourceUnstructured(*createdUn, resType), nil
}

func (c *Resources) Update(resource Resource) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { fmt.Printf("update %s\n", time.Now().UTC().Sub(t1)) }()

		bs, _ := resource.AsYAMLBytes()
		fmt.Printf("update resource %s\n%s\n", resource.Description(), bs)
	}

	resClient, resType, err := c.resourceClient(resource)
	if err != nil {
		return nil, err
	}

	var updatedUn *unstructured.Unstructured

	err = util.Retry(time.Second, time.Minute, func() (bool, error) {
		updatedUn, err = resClient.Update(resource.unstructuredPtr())
		if err != nil {
			return c.doneRetryingErr(err), c.resourceErr(err, "Updating", resource)
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return NewResourceUnstructured(*updatedUn, resType), nil
}

func (c *Resources) Patch(resource Resource, patchType types.PatchType, data []byte) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { fmt.Printf("patch %s\n", time.Now().UTC().Sub(t1)) }()
	}

	resClient, resType, err := c.resourceClient(resource)
	if err != nil {
		return nil, err
	}

	var patchedUn *unstructured.Unstructured

	err = util.Retry(time.Second, time.Minute, func() (bool, error) {
		patchedUn, err = resClient.Patch(resource.Name(), patchType, data)
		if err != nil {
			return c.doneRetryingErr(err), c.resourceErr(err, "Patching", resource)
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return NewResourceUnstructured(*patchedUn, resType), nil
}

const (
	admissionWebhookErrMsg = "Internal error occurred: failed calling admission webhook"
)

func (c *Resources) doneRetryingErr(err error) bool {
	// TODO is there a better way to detect this error
	retry := strings.Contains(err.Error(), admissionWebhookErrMsg)
	return !retry
}

func (c *Resources) Delete(resource Resource) error {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { fmt.Printf("delete %s\n", time.Now().UTC().Sub(t1)) }()
	}

	if resource.IsDeleting() {
		fmt.Printf("TODO resource '%s' is already deleting\n", resource.Description())
		return nil
	}

	resClient, resType, err := c.resourceClient(resource)
	if err != nil {
		return err
	}

	if resType.Deletable() {
		// TODO is setting deletion policy a correct thing to do?
		// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#setting-the-cascading-deletion-policy
		delPol := metav1.DeletePropagationBackground
		delOpts := &metav1.DeleteOptions{PropagationPolicy: &delPol}

		// Some resources may not have UID (example: PodMetrics.metrics.k8s.io)
		resUID := types.UID(resource.UID())
		if len(resUID) > 0 {
			delOpts.Preconditions = &metav1.Preconditions{UID: &resUID}
		}

		err = resClient.Delete(resource.Name(), delOpts)
		if err != nil {
			if errors.IsNotFound(err) || strings.Contains(err.Error(), "not found") { // TODO why "not found" check is needed?
				fmt.Printf("TODO resource is not found: %s (reason: %s)\n", resource.Description(), errors.ReasonForError(err))
				return nil
			}
			return c.resourceErr(err, "Deleting", resource)
		}
	} else {
		fmt.Printf("TODO resource is not deletable: %s\n", resource.Description()) // TODO
	}

	return nil
}

func (c *Resources) Get(resource Resource) (Resource, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { fmt.Printf("get %s\n", time.Now().UTC().Sub(t1)) }()
	}

	resClient, resType, err := c.resourceClient(resource)
	if err != nil {
		return nil, err
	}

	item, err := resClient.Get(resource.Name(), metav1.GetOptions{})
	if err != nil {
		return nil, c.resourceErr(err, "Getting", resource)
	}

	return NewResourceUnstructured(*item, resType), nil
}

func (c *Resources) Exists(resource Resource) (bool, error) {
	if resourcesDebug {
		t1 := time.Now().UTC()
		defer func() { fmt.Printf("exists %s\n", time.Now().UTC().Sub(t1)) }()
	}

	resClient, resType, err := c.resourceClient(resource)
	if err != nil {
		// Assume if type is not known to the API server
		// then such resource cannot exist on the server
		if _, ok := err.(ResourceTypesUnknownTypeErr); ok {
			return false, nil
		}
		return false, err
	}

	var found bool

	err = util.Retry(time.Second, time.Minute, func() (bool, error) {
		_, err = resClient.Get(resource.Name(), metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				found = false
				return true, nil
			}
			// Abnormal error case. Note that it says pod is not found even though we were checking on podmetrics.
			// `Checking existance of resource podmetrics/knative-ingressgateway-646d475cbb-c82qb (metrics.k8s.io/v1beta1)
			// namespace: istio-system: Error while getting pod knative-ingressgateway-646d475cbb-c82qb:
			// pod "knative-ingressgateway-646d475cbb-c82qb" not found (reason: )`
			if strings.Contains(err.Error(), "not found") {
				found, err = c.expensiveExistsViaList(resType, resource)
				if err != nil {
					err = c.resourceErr(err, "Checking existance (expensive) of", resource)
				}
				return true, err
			}
			// TODO sometimes metav1.StatusReasonUnknown is returned (empty string)
			// might be related to deletion of mutating webhook
			return false, c.resourceErr(err, "Checking existance of", resource)
		}

		found = true
		return true, nil
	})

	return found, err
}

func (c *Resources) resourceErr(err error, action string, resource Resource) error {
	if typedErr, ok := err.(errors.APIStatus); ok {
		return resourceStatusErr{resourcePlainErr{err, action, resource}, typedErr.Status()}
	}
	return resourcePlainErr{err, action, resource}
}

func (c *Resources) resourceClient(resource Resource) (dynamic.ResourceInterface, ResourceType, error) {
	resType, err := c.resourceTypes.Find(resource)
	if err != nil {
		return nil, ResourceType{}, err
	}

	return c.dynamicClient.Resource(resType.GroupVersionResource).Namespace(resource.Namespace()), resType, nil
}

func (c *Resources) expensiveExistsViaList(resType ResourceType, resource Resource) (bool, error) {
	rs, err := c.All([]ResourceType{resType}, ResourcesAllOpts{})
	if err != nil {
		return false, err
	}

	// Use UniqueResourceKey instead of UID as UID may not be set (example: metrics.k8s.io/PodMetrics)
	resourceKey := NewUniqueResourceKey(resource).String()

	for _, res := range rs {
		resKey := NewUniqueResourceKey(res).String()
		if resKey == resourceKey {
			return true, nil
		}
	}

	return false, nil
}

func (c *Resources) assumedAllowedNamespaces() ([]string, error) {
	c.assumedAllowedNamespacesMemoLock.Lock()
	defer c.assumedAllowedNamespacesMemoLock.Unlock()

	if c.assumedAllowedNamespacesMemo != nil {
		return *c.assumedAllowedNamespacesMemo, nil
	}

	nsList, err := c.coreClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		if errors.IsForbidden(err) {
			if len(c.fallbackAllowedNamespaces) > 0 {
				return c.fallbackAllowedNamespaces, nil
			}
		}
		return nil, fmt.Errorf("fetching all namespaces: %s", err)
	}

	var nsNames []string

	for _, ns := range nsList.Items {
		nsNames = append(nsNames, ns.Name)
	}

	c.assumedAllowedNamespacesMemo = &nsNames

	return nsNames, nil
}

type ResourcesAllOpts struct {
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
	return fmt.Sprintf("%s resource %s: %s (reason: %s)",
		e.action, e.resource.Description(), e.err, errors.ReasonForError(e.err))
}
