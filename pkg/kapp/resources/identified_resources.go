package resources

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type IdentifiedResources struct {
	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	resourceTypes ResourceTypes
	resources     Resources
}

func NewIdentifiedResources(coreClient kubernetes.Interface, dynamicClient dynamic.Interface) IdentifiedResources {
	resTypes := NewResourceTypesImpl(coreClient)
	return IdentifiedResources{
		coreClient,
		dynamicClient,
		resTypes,
		NewResources(resTypes, coreClient, dynamicClient),
	}
}

func (r IdentifiedResources) Create(resource Resource) (Resource, error) {
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
	return r.resources.Patch(resource, patchType, data)
}

func (r IdentifiedResources) Delete(resource Resource) error {
	return r.resources.Delete(resource)
}

func (r IdentifiedResources) Get(resource Resource) (Resource, error) {
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
	return r.resources.Exists(resource)
}
