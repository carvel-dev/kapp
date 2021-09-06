// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// +k8s:openapi-gen=true
type AppDeploy struct {
	Kapp *AppDeployKapp `json:"kapp,omitempty" protobuf:"bytes,1,opt,name=kapp"`
}

// +k8s:openapi-gen=true
type AppDeployKapp struct {
	IntoNs     string   `json:"intoNs,omitempty" protobuf:"bytes,1,opt,name=intoNs"`
	MapNs      []string `json:"mapNs,omitempty" protobuf:"bytes,2,rep,name=mapNs"`
	RawOptions []string `json:"rawOptions,omitempty" protobuf:"bytes,3,rep,name=rawOptions"`

	Inspect *AppDeployKappInspect `json:"inspect,omitempty" protobuf:"bytes,4,opt,name=inspect"`
	Delete  *AppDeployKappDelete  `json:"delete,omitempty" protobuf:"bytes,5,opt,name=delete"`
}

// +k8s:openapi-gen=true
type AppDeployKappInspect struct {
	RawOptions []string `json:"rawOptions,omitempty" protobuf:"bytes,1,rep,name=rawOptions"`
}

// +k8s:openapi-gen=true
type AppDeployKappDelete struct {
	RawOptions []string `json:"rawOptions,omitempty" protobuf:"bytes,1,rep,name=rawOptions"`
}
