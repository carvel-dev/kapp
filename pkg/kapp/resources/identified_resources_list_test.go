// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package resources

import (
	"testing"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestIdentifiedResourcesListReturnsLabeledResources(t *testing.T) {
	fakeResourceTypes := &FakeResourceTypes{}
	fakeResources := &FakeResources{t}

	identifiedResources := NewIdentifiedResources(nil, fakeResourceTypes, fakeResources, []string{}, logger.NewUILogger(ui.NewNoopUI()))
	sel := labels.Set(map[string]string{"some-label": "value"}).AsSelector()

	resources, err := identifiedResources.List(sel, nil)
	require.Nil(t, err)
	require.NotNil(t, resources)

	require.Equal(t, 1, len(resources))
	require.Contains(t, resources[0].Labels(), "some-label")
	require.Equal(t, resources[0].Labels()["some-label"], "value")
}

type FakeResources struct {
	t *testing.T
}

func (r *FakeResources) All([]ResourceType, AllOpts) ([]Resource, error) {
	antreaBs := `---
apiVersion: clusterinformation.antrea.tanzu.vmware.com/v1beta1
kind: AntreaControllerInfo
metadata:
  name: antrea-controller
version: v0.10.1
`

	deploymentBs := `---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    some-label: "value"
`

	antreaRes := MustNewResourceFromBytes([]byte(antreaBs))
	deploymentRes := MustNewResourceFromBytes([]byte(deploymentBs))

	return []Resource{antreaRes, deploymentRes}, nil
}
func (r *FakeResources) Delete(Resource) error                                     { return nil }
func (r *FakeResources) Exists(Resource) (bool, error)                             { return true, nil }
func (r *FakeResources) Get(Resource) (Resource, error)                            { return nil, nil }
func (r *FakeResources) Patch(Resource, types.PatchType, []byte) (Resource, error) { return nil, nil }
func (r *FakeResources) Update(Resource) (Resource, error)                         { return nil, nil }
func (r *FakeResources) Create(Resource) (Resource, error)                         { return nil, nil }

type FakeResourceTypes struct{}

func (r *FakeResourceTypes) All() ([]ResourceType, error)                          { return nil, nil }
func (r *FakeResourceTypes) Find(Resource) (ResourceType, error)                   { return ResourceType{}, nil }
func (r *FakeResourceTypes) CanIgnoreFailingGroupVersion(schema.GroupVersion) bool { return true }
