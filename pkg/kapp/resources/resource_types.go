// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

type ResourceTypes interface {
	All(ignoreCachedResTypes bool) ([]ResourceType, error)
	Find(Resource) (ResourceType, error)
	CanIgnoreFailingGroupVersion(schema.GroupVersion) bool
}

type ResourceTypesImplOpts struct {
	IgnoreFailingAPIServices   bool
	CanIgnoreFailingAPIService func(schema.GroupVersion) bool
}

type ResourceTypesImpl struct {
	coreClient kubernetes.Interface
	opts       ResourceTypesImplOpts

	memoizedResTypes     *[]ResourceType
	memoizedResTypesLock sync.RWMutex
}

var _ ResourceTypes = &ResourceTypesImpl{}

type ResourceType struct {
	schema.GroupVersionResource
	metav1.APIResource
}

func NewResourceTypesImpl(coreClient kubernetes.Interface, opts ResourceTypesImplOpts) *ResourceTypesImpl {
	return &ResourceTypesImpl{coreClient: coreClient, opts: opts}
}

func (g *ResourceTypesImpl) All(ignoreCachedResTypes bool) ([]ResourceType, error) {
	if ignoreCachedResTypes {
		// TODO Update cache while doing a fresh fetch
		return g.all()
	}
	return g.memoizedAll()
}

func (g *ResourceTypesImpl) all() ([]ResourceType, error) {
	serverResources, err := g.serverResources()
	if err != nil {
		return nil, err
	}

	var pairs []ResourceType

	for _, resList := range serverResources {
		groupVersion, err := schema.ParseGroupVersion(resList.GroupVersion)
		if err != nil {
			return nil, err
		}

		for _, res := range resList.APIResources {
			group := groupVersion.Group
			if len(res.Group) > 0 {
				group = res.Group
			}

			version := groupVersion.Version
			if len(res.Version) > 0 {
				version = res.Version
			}

			// Copy down group and version for convenience
			res.Group = group
			res.Version = version

			gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: res.Name}
			pairs = append(pairs, ResourceType{gvr, res})
		}
	}

	return pairs, nil
}

func (g *ResourceTypesImpl) CanIgnoreFailingGroupVersion(groupVer schema.GroupVersion) bool {
	return g.canIgnoreFailingGroupVersions(map[schema.GroupVersion]error{groupVer: nil})
}

func (g *ResourceTypesImpl) canIgnoreFailingGroupVersions(groupVers map[schema.GroupVersion]error) bool {
	// If groups that are failing do not relate to our resources
	// it's ok to ignore them. Still not ideal but not much else
	// we can do with the way kubernetes exposes this functionality.
	if g.opts.IgnoreFailingAPIServices {
		return true
	}
	if g.opts.CanIgnoreFailingAPIService != nil {
		for groupVer := range groupVers {
			if !g.opts.CanIgnoreFailingAPIService(groupVer) {
				return false
			}
		}
		return true
	}
	return false
}

func (g *ResourceTypesImpl) serverResources() ([]*metav1.APIResourceList, error) {
	var serverResources []*metav1.APIResourceList
	var lastErr error

	for i := 0; i < 10; i++ {
		serverResources, lastErr = g.coreClient.Discovery().ServerResources()
		if lastErr == nil {
			return serverResources, nil
		} else if typedLastErr, ok := lastErr.(*discovery.ErrGroupDiscoveryFailed); ok {
			if len(serverResources) > 0 && g.canIgnoreFailingGroupVersions(typedLastErr.Groups) {
				return serverResources, nil
			}
			// Even local services may not be Available immediately, so retry
			lastErr = fmt.Errorf("%s (possibly related issue: https://github.com/vmware-tanzu/carvel-kapp/issues/12)", lastErr)
		}
		time.Sleep(1 * time.Second)
	}

	return nil, lastErr
}

func (g *ResourceTypesImpl) memoizedAll() ([]ResourceType, error) {
	g.memoizedResTypesLock.RLock()

	if g.memoizedResTypes != nil {
		defer g.memoizedResTypesLock.RUnlock()
		return *g.memoizedResTypes, nil
	}

	g.memoizedResTypesLock.RUnlock()

	// Include call to All within a lock to avoid race
	// with competing memoizedAll() call that
	// may win and save older copy on res types
	g.memoizedResTypesLock.Lock()
	defer g.memoizedResTypesLock.Unlock()

	resTypes, err := g.all()
	if err != nil {
		return nil, err
	}

	g.memoizedResTypes = &resTypes
	return resTypes, nil
}

func (g *ResourceTypesImpl) Find(resource Resource) (ResourceType, error) {
	resType, err := g.findOnce(resource)
	if err != nil {
		g.memoizedResTypesLock.Lock()
		g.memoizedResTypes = nil
		g.memoizedResTypesLock.Unlock()

		return g.findOnce(resource)
	}

	return resType, nil
}

type ResourceTypesUnknownTypeErr struct {
	resource Resource
}

func (e ResourceTypesUnknownTypeErr) Error() string {
	return "Expected to find type for resource: " + e.resource.Description()
}

func (g *ResourceTypesImpl) findOnce(resource Resource) (ResourceType, error) {
	pairs, err := g.memoizedAll()
	if err != nil {
		return ResourceType{}, err
	}

	pieces := strings.Split(resource.APIVersion(), "/")
	if len(pieces) > 2 {
		return ResourceType{}, fmt.Errorf("Expected version to be of format group/version")
	}
	if len(pieces) == 1 {
		pieces = []string{"", pieces[0]} // core API group
	}

	for _, pair := range pairs {
		if pair.APIResource.Group == pieces[0] &&
			pair.APIResource.Version == pieces[1] &&
			pair.APIResource.Kind == resource.Kind() {
			return pair, nil
		}
	}

	return ResourceType{}, ResourceTypesUnknownTypeErr{resource}
}

func (p ResourceType) Namespaced() bool {
	return p.APIResource.Namespaced
}

func (p ResourceType) Listable() bool {
	return p.containsStr(p.APIResource.Verbs, "list")
}

func (p ResourceType) Deletable() bool {
	return p.containsStr(p.APIResource.Verbs, "delete")
}

func (p ResourceType) containsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Listable(in []ResourceType) []ResourceType {
	var out []ResourceType
	for _, item := range in {
		if item.Listable() {
			out = append(out, item)
		}
	}
	return out
}

func Matching(in []ResourceType, ref ResourceRef) []ResourceType {
	partResourceRef := PartialResourceRef{ref.GroupVersionResource}
	var out []ResourceType
	for _, item := range in {
		if partResourceRef.Matches(item.GroupVersionResource) {
			out = append(out, item)
		}
	}
	return out
}

func MatchingAny(in []ResourceType, refs []ResourceRef) []ResourceType {
	var out []ResourceType
	for _, item := range in {
		for _, ref := range refs {
			if (PartialResourceRef{ref.GroupVersionResource}).Matches(item.GroupVersionResource) {
				out = append(out, item)
				break
			}
		}
	}
	return out
}

func NonMatching(in []ResourceType, ref ResourceRef) []ResourceType {
	partResourceRef := PartialResourceRef{ref.GroupVersionResource}
	var out []ResourceType
	for _, item := range in {
		if !partResourceRef.Matches(item.GroupVersionResource) {
			out = append(out, item)
		}
	}
	return out
}

// TODO: Extend ResourceRef and PartialResourceRefd to allow GVK matching
func MatchingAnyGK(in []ResourceType, gks []schema.GroupKind) []ResourceType {
	var out []ResourceType
	for _, item := range in {
		for _, gk := range gks {
			if (GKResourceRef{gk}).Matches(item) {
				out = append(out, item)
			}
		}
	}
	return out
}
