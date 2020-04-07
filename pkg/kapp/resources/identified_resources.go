package resources

import (
	"fmt"

	"github.com/k14s/kapp/pkg/kapp/logger"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type IdentifiedResources struct {
	coreClient                kubernetes.Interface
	fallbackAllowedNamespaces []string
	resourceTypes             ResourceTypes
	resources                 *Resources
	logger                    logger.Logger
}

func NewIdentifiedResources(coreClient kubernetes.Interface,
	dynamicClient dynamic.Interface, resourceTypes ResourceTypes,
	fallbackAllowedNamespaces []string, logger logger.Logger) IdentifiedResources {

	resources := NewResources(resourceTypes, coreClient, dynamicClient, fallbackAllowedNamespaces, logger)

	return IdentifiedResources{coreClient, fallbackAllowedNamespaces, resourceTypes, resources, logger.NewPrefixed("IdentifiedResources")}
}

func (r IdentifiedResources) Create(resource Resource) (Resource, error) {
	defer r.logger.DebugFunc(fmt.Sprintf("Create(%s)", resource.Description())).Finish()

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
	defer r.logger.DebugFunc(fmt.Sprintf("Update(%s)", resource.Description())).Finish()

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
	defer r.logger.DebugFunc(fmt.Sprintf("Patch(%s)", resource.Description())).Finish()
	return r.resources.Patch(resource, patchType, data)
}

func (r IdentifiedResources) Delete(resource Resource) error {
	defer r.logger.DebugFunc(fmt.Sprintf("Delete(%s)", resource.Description())).Finish()
	return r.resources.Delete(resource)
}

func (r IdentifiedResources) Get(resource Resource) (Resource, error) {
	defer r.logger.DebugFunc(fmt.Sprintf("Get(%s)", resource.Description())).Finish()

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
	defer r.logger.DebugFunc(fmt.Sprintf("Exists(%s)", resource.Description())).Finish()
	return r.resources.Exists(resource)
}
