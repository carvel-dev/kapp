package resources

import (
	"github.com/k14s/kapp/pkg/kapp/logger"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type IdentifiedResources struct {
	coreClient    kubernetes.Interface
	resourceTypes ResourceTypes
	resources     *Resources
	logger        logger.Logger
}

func NewIdentifiedResources(coreClient kubernetes.Interface,
	dynamicClient dynamic.Interface, fallbackAllowedNamespaces []string, logger logger.Logger) IdentifiedResources {

	resTypes := NewResourceTypesImpl(coreClient)
	resources := NewResources(resTypes, coreClient, dynamicClient, fallbackAllowedNamespaces)

	return IdentifiedResources{coreClient, resTypes, resources, logger}
}

func (r IdentifiedResources) Create(resource Resource) (Resource, error) {
	defer r.logger.DebugFunc("Create").Finish()

	resource = resource.DeepCopy()

	err := NewIdentityAnnotation(resource).AddMod().Apply(resource)
	if err != nil {
		return nil, err
	}

	resource, err = r.resources.Create(resource)
	if err != nil {
		return nil, err
	}

	err = NewIdentityAnnotation(resource).RemoveMod().Apply(resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r IdentifiedResources) Update(resource Resource) (Resource, error) {
	defer r.logger.DebugFunc("Update").Finish()

	resource = resource.DeepCopy()

	err := NewIdentityAnnotation(resource).AddMod().Apply(resource)
	if err != nil {
		return nil, err
	}

	resource, err = r.resources.Update(resource)
	if err != nil {
		return nil, err
	}

	err = NewIdentityAnnotation(resource).RemoveMod().Apply(resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r IdentifiedResources) Patch(resource Resource, patchType types.PatchType, data []byte) (Resource, error) {
	defer r.logger.DebugFunc("Patch").Finish()
	return r.resources.Patch(resource, patchType, data)
}

func (r IdentifiedResources) Delete(resource Resource) error {
	defer r.logger.DebugFunc("Delete").Finish()
	return r.resources.Delete(resource)
}

func (r IdentifiedResources) Get(resource Resource) (Resource, error) {
	defer r.logger.DebugFunc("Get").Finish()

	resource, err := r.resources.Get(resource)
	if err != nil {
		return nil, err
	}

	err = NewIdentityAnnotation(resource).RemoveMod().Apply(resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r IdentifiedResources) Exists(resource Resource) (bool, error) {
	defer r.logger.DebugFunc("Exists").Finish()
	return r.resources.Exists(resource)
}
