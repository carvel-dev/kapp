package resourcesmisc

import (
	"fmt"
	"strings"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type ResourceTypes struct {
	localCRDs      []*ApiExtensionsVxCRD
	resourceTypes  ctlres.ResourceTypes
	memoizedScopes map[string]bool
}

func NewResourceTypes(newResources []ctlres.Resource, coreClient kubernetes.Interface, dynamicClient dynamic.Interface) *ResourceTypes {
	var localCRDs []*ApiExtensionsVxCRD

	for _, newRes := range newResources {
		crd := NewApiExtensionsVxCRD(newRes)
		if crd != nil {
			localCRDs = append(localCRDs, crd)
		}
	}

	return &ResourceTypes{localCRDs, ctlres.NewResourceTypesImpl(coreClient), nil}
}

func (c *ResourceTypes) IsNamespaced(res ctlres.Resource) (bool, error) {
	scopeMap, err := c.scopeMap()
	if err != nil {
		return false, err
	}

	apiVer := res.APIVersion()
	if !strings.Contains(apiVer, "/") {
		apiVer = "/" + apiVer // core group is empty
	}

	fullKind := apiVer + "/" + res.Kind()

	isNamespaced, found := scopeMap[fullKind]
	if !found {
		return false, fmt.Errorf("Expected to find kind '%s', but did not", fullKind)
	}

	return isNamespaced, nil
}

func (c *ResourceTypes) scopeMap() (map[string]bool, error) {
	if c.memoizedScopes != nil {
		return c.memoizedScopes, nil
	}

	scopeMap, err := c.clusterScopes()
	if err != nil {
		return nil, err
	}

	scopeMap2, err := c.localCRDScopes()
	if err != nil {
		return nil, err
	}

	// Additional CRDs last to override cluster config
	for k, v := range scopeMap2 {
		scopeMap[k] = v
	}

	c.memoizedScopes = scopeMap

	return scopeMap, nil
}

func (c *ResourceTypes) clusterScopes() (map[string]bool, error) {
	scopeMap := map[string]bool{}

	resTypes, err := c.resourceTypes.All()
	if err != nil {
		return nil, err
	}

	for _, resType := range resTypes {
		key := resType.APIResource.Group + "/" + resType.APIResource.Version + "/" + resType.APIResource.Kind
		scopeMap[key] = resType.APIResource.Namespaced
	}

	return scopeMap, nil
}

func (c *ResourceTypes) localCRDScopes() (map[string]bool, error) {
	scopeMap := map[string]bool{}

	for _, crd := range c.localCRDs {
		contents, err := crd.contents()
		if err != nil {
			return nil, err
		}

		for _, ver := range contents.Versions() {
			key := contents.Spec.Group + "/" + ver + "/" + contents.Spec.Names.Kind
			scopeMap[key] = contents.Spec.Scope == "Namespaced"
		}
	}

	return scopeMap, nil
}
