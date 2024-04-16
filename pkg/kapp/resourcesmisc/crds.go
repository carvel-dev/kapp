// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"
	"strings"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
)

type ResourceTypes struct {
	localCRDs      []*APIExtensionsVxCRD
	resourceTypes  ctlres.ResourceTypes
	memoizedScopes map[string]bool
}

func NewResourceTypes(newResources []ctlres.Resource, resourceTypes ctlres.ResourceTypes) *ResourceTypes {
	var localCRDs []*APIExtensionsVxCRD

	for _, newRes := range newResources {
		crd := NewAPIExtensionsVxCRD(newRes)
		if crd != nil {
			localCRDs = append(localCRDs, crd)
		}
	}

	return &ResourceTypes{localCRDs, resourceTypes, nil}
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
		msgs := []string{
			"- Kubernetes API server did not have matching apiVersion + kind",
			"- No matching CRD was found in given configuration",
		}
		return false, fmt.Errorf("Expected to find kind '%s', but did not:\n%s", fullKind, strings.Join(msgs, "\n"))
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

	resTypes, err := c.resourceTypes.All(false)
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
