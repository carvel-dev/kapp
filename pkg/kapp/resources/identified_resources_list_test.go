// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
package resources

import (
	"github.com/cppforlife/go-cli-ui/ui"
	logger2 "github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"testing"

	"k8s.io/apimachinery/pkg/labels"
)

func TestIdentifiedResourcesListReturnsLabeledResources(t *testing.T) {
	mockedResTypes := new(ResourceTypesMock)
	mockedRess := new(ResourcesMock)

	mockedRess.Mock.On("All").Return(makeNewResources(t), nil)

	ui := ui.NewNoopUI()
	logger := logger2.NewUILogger(ui)
	identifiedResources := NewIdentifiedResources(nil, mockedResTypes, mockedRess, []string{}, logger)

	kappLabel := make(map[string]string)
	kappLabel["kapp.k14s.io/app"] = "app-name"
	sel := labels.Set(kappLabel).AsSelector()

	resources, err := identifiedResources.List(sel, nil)

	mockedRess.AssertCalled(t, "All")
	require.Nil(t, err)
	require.NotNil(t, resources)

	for _, res := range resources {
		require.Contains(t, res.Labels(), "kapp.k14s.io/app")
		require.Equal(t, res.Labels()["kapp.k14s.io/app"], kappLabel["kapp.k14s.io/app"])
	}
}

type ResourcesMock struct {
	mock.Mock
}

func (r *ResourcesMock) All([]ResourceType, AllOpts) ([]Resource, error) {
	args := r.Called()
	return args.Get(0).([]Resource), args.Error(1)
}
func (r *ResourcesMock) Delete(Resource) error                                     { return nil }
func (r *ResourcesMock) Exists(Resource) (bool, error)                             { return true, nil }
func (r *ResourcesMock) Get(Resource) (Resource, error)                            { return nil, nil }
func (r *ResourcesMock) Patch(Resource, types.PatchType, []byte) (Resource, error) { return nil, nil }
func (r *ResourcesMock) Update(Resource) (Resource, error)                         { return nil, nil }
func (r *ResourcesMock) Create(Resource) (Resource, error)                         { return nil, nil }

type ResourceTypesMock struct {
	mock.Mock
}

func (r *ResourceTypesMock) All() ([]ResourceType, error)                          { return nil, nil }
func (r *ResourceTypesMock) Find(Resource) (ResourceType, error)                   { return ResourceType{}, nil }
func (r *ResourceTypesMock) CanIgnoreFailingGroupVersion(schema.GroupVersion) bool { return true }

func makeNewResources(t *testing.T) []Resource {
	t.Helper()
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
  annotations:
    deployment.kubernetes.io/revision: "1"
    kapp.k14s.io/identity: v1;default/apps/Deployment/nginx-deployment;apps/v1
    kapp.k14s.io/original: '{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"labels":{"app":"nginx","kapp.k14s.io/app":"1614279362730868000","kapp.k14s.io/association":"v1.5771d3c49b880a055e733831a44ae242"},"name":"nginx-deployment","namespace":"default"},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"nginx","kapp.k14s.io/app":"1614279362730868000"}},"template":{"metadata":{"labels":{"app":"nginx","kapp.k14s.io/app":"1614279362730868000","kapp.k14s.io/association":"v1.5771d3c49b880a055e733831a44ae242"}},"spec":{"containers":[{"image":"nginx:1.14.2","name":"nginx","ports":[{"containerPort":80}]}]}}}}'
    kapp.k14s.io/original-diff-md5: 2b7269146768d693cb97b932660a532d
  labels:
    app: nginx
    kapp.k14s.io/app: "app-name"
    kapp.k14s.io/association: v1.5771d3c49b880a055e733831a44ae242
  name: nginx-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
      kapp.k14s.io/app: "1614279362730868000"
  template:
    metadata:
      labels:
        app: nginx
        kapp.k14s.io/app: "1614279362730868000"
        kapp.k14s.io/association: v1.5771d3c49b880a055e733831a44ae242
    spec:
      containers:
      - image: nginx:1.14.2
        name: nginx
        ports:
        - containerPort: 80
          protocol: TCP
`

	antreaRes, err := NewResourceFromBytes([]byte(antreaBs))
	require.Nil(t, err)
	deploymentRes, err := NewResourceFromBytes([]byte(deploymentBs))
	require.Nil(t, err)

	return []Resource{antreaRes, deploymentRes}
}
