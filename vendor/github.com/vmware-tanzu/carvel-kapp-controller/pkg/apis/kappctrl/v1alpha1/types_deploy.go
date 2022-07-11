// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// +k8s:openapi-gen=true
type AppDeploy struct {
	// Use kapp to deploy resources
	Kapp *AppDeployKapp `json:"kapp,omitempty" protobuf:"bytes,1,opt,name=kapp"`
}

// +k8s:openapi-gen=true
type AppDeployKapp struct {
	// Override namespace for all resources (optional)
	IntoNs string `json:"intoNs,omitempty" protobuf:"bytes,1,opt,name=intoNs"`
	// Provide custom namespace override mapping (optional)
	MapNs []string `json:"mapNs,omitempty" protobuf:"bytes,2,rep,name=mapNs"`
	// Pass through options to kapp deploy (optional)
	RawOptions []string `json:"rawOptions,omitempty" protobuf:"bytes,3,rep,name=rawOptions"`

	// Configuration for inspect command (optional)
	// as of kapp-controller v0.31.0, inspect is disabled by default
	// add rawOptions or use an empty inspect config like `inspect: {}` to enable
	Inspect *AppDeployKappInspect `json:"inspect,omitempty" protobuf:"bytes,4,opt,name=inspect"`
	// Configuration for delete command (optional)
	Delete *AppDeployKappDelete `json:"delete,omitempty" protobuf:"bytes,5,opt,name=delete"`
}

// +k8s:openapi-gen=true
type AppDeployKappInspect struct {
	// Pass through options to kapp inspect (optional)
	RawOptions []string `json:"rawOptions,omitempty" protobuf:"bytes,1,rep,name=rawOptions"`
}

// +k8s:openapi-gen=true
type AppDeployKappDelete struct {
	// Pass through options to kapp delete (optional)
	RawOptions []string `json:"rawOptions,omitempty" protobuf:"bytes,1,rep,name=rawOptions"`
}
