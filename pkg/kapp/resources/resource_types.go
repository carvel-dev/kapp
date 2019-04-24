package resources

import (
	"fmt"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type ResourceTypes interface {
	All() ([]ResourceType, error)
	Find(Resource) (ResourceType, error)
}

type ResourceTypesImpl struct {
	coreClient           kubernetes.Interface
	memoizedResTypes     *[]ResourceType
	memoizedResTypesLock sync.RWMutex
}

var _ ResourceTypes = &ResourceTypesImpl{}

type ResourceType struct {
	schema.GroupVersionResource
	metav1.APIResource
}

func NewResourceTypesImpl(coreClient kubernetes.Interface) *ResourceTypesImpl {
	return &ResourceTypesImpl{coreClient: coreClient}
}

func (g *ResourceTypesImpl) All() ([]ResourceType, error) {
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

func (g *ResourceTypesImpl) serverResources() ([]*metav1.APIResourceList, error) {
	var lastError error
	for i := 0; i < 5; i++ {
		serverResources, err := g.coreClient.Discovery().ServerResources()
		if err == nil {
			return serverResources, nil
		}
		lastError = err
		time.Sleep(1 * time.Second)
	}
	return nil, lastError
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

	resTypes, err := g.All()
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
