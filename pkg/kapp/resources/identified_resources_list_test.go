// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package resources_test

import (
	"testing"

	"carvel.dev/kapp/pkg/kapp/logger"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestIdentifiedResourcesListReturnsLabeledResources(t *testing.T) {
	fakeResourceTypes := &FakeResourceTypes{}
	fakeResources := &FakeResources{t}

	identifiedResources := ctlres.NewIdentifiedResources(nil, fakeResourceTypes, fakeResources, []string{}, logger.NewUILogger(ui.NewNoopUI()))
	sel := labels.Set(map[string]string{"some-label": "value"}).AsSelector()

	resources, err := identifiedResources.List(sel, nil, ctlres.IdentifiedResourcesListOpts{})
	require.Nil(t, err)
	require.NotNil(t, resources)

	require.Equal(t, 1, len(resources))
	require.Contains(t, resources[0].Labels(), "some-label")
	require.Equal(t, resources[0].Labels()["some-label"], "value")
}

type FakeResources struct {
	t *testing.T
}

func (r *FakeResources) All([]ctlres.ResourceType, ctlres.AllOpts) ([]ctlres.Resource, error) {
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

	antreaRes := ctlres.MustNewResourceFromBytes([]byte(antreaBs))
	deploymentRes := ctlres.MustNewResourceFromBytes([]byte(deploymentBs))

	return []ctlres.Resource{antreaRes, deploymentRes}, nil
}
func (r *FakeResources) Delete(ctlres.Resource) error { return nil }
func (r *FakeResources) Exists(ctlres.Resource, ctlres.ExistsOpts) (ctlres.Resource, bool, error) {
	return nil, true, nil
}
func (r *FakeResources) Get(ctlres.Resource) (ctlres.Resource, error) { return nil, nil }
func (r *FakeResources) Patch(ctlres.Resource, types.PatchType, []byte) (ctlres.Resource, error) {
	return nil, nil
}
func (r *FakeResources) Update(ctlres.Resource) (ctlres.Resource, error) { return nil, nil }
func (r *FakeResources) Create(ctlres.Resource) (ctlres.Resource, error) { return nil, nil }

type FakeResourceTypes struct{}

func (r *FakeResourceTypes) All(_ bool) ([]ctlres.ResourceType, error) {
	return nil, nil
}
func (r *FakeResourceTypes) Find(ctlres.Resource) (ctlres.ResourceType, error) {
	return ctlres.ResourceType{}, nil
}
func (r *FakeResourceTypes) CanIgnoreFailingGroupVersion(schema.GroupVersion) bool { return true }
