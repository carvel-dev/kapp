// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
)

// +k8s:openapi-gen=true
type AppFetch struct {
	Inline       *AppFetchInline       `json:"inline,omitempty" protobuf:"bytes,1,opt,name=inline"`
	Image        *AppFetchImage        `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	HTTP         *AppFetchHTTP         `json:"http,omitempty" protobuf:"bytes,3,opt,name=http"`
	Git          *AppFetchGit          `json:"git,omitempty" protobuf:"bytes,4,opt,name=git"`
	HelmChart    *AppFetchHelmChart    `json:"helmChart,omitempty" protobuf:"bytes,5,opt,name=helmChart"`
	ImgpkgBundle *AppFetchImgpkgBundle `json:"imgpkgBundle,omitempty" protobuf:"bytes,6,opt,name=imgpkgBundle"`
}

// +k8s:openapi-gen=true
type AppFetchInline struct {
	Paths     map[string]string      `json:"paths,omitempty" protobuf:"bytes,1,rep,name=paths"`
	PathsFrom []AppFetchInlineSource `json:"pathsFrom,omitempty" protobuf:"bytes,2,rep,name=pathsFrom"`
}

// +k8s:openapi-gen=true
type AppFetchInlineSource struct {
	SecretRef    *AppFetchInlineSourceRef `json:"secretRef,omitempty" protobuf:"bytes,1,opt,name=secretRef"`
	ConfigMapRef *AppFetchInlineSourceRef `json:"configMapRef,omitempty" protobuf:"bytes,2,opt,name=configMapRef"`
}

// +k8s:openapi-gen=true
type AppFetchInlineSourceRef struct {
	DirectoryPath string `json:"directoryPath,omitempty" protobuf:"bytes,2,opt,name=directoryPath"`
	Name          string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// +k8s:openapi-gen=true
type AppFetchImage struct {
	// Example: username/app1-config:v0.1.0
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// +optional
	TagSelection *versions.VersionSelection `json:"tagSelection,omitempty" protobuf:"bytes,4,opt,name=tagSelection"`
	// Secret may include one or more keys: username, password, token.
	// By default anonymous access is used for authentication.
	// TODO support docker config formated secret
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,2,opt,name=secretRef"`
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,3,opt,name=subPath"`
}

// +k8s:openapi-gen=true
type AppFetchHTTP struct {
	// URL can point to one of following formats: text, tgz, zip
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// +optional
	SHA256 string `json:"sha256,omitempty" protobuf:"bytes,2,opt,name=sha256"`
	// Secret may include one or more keys: username, password
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,3,opt,name=secretRef"`
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`
}

// TODO implement git
// +k8s:openapi-gen=true
type AppFetchGit struct {
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// +optional
	Ref string `json:"ref,omitempty" protobuf:"bytes,2,opt,name=ref"`
	// +optional
	RefSelection *versions.VersionSelection `json:"refSelection,omitempty" protobuf:"bytes,6,opt,name=refSelection"`
	// Secret may include one or more keys: ssh-privatekey, ssh-knownhosts
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,3,opt,name=secretRef"`
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`
	// +optional
	LFSSkipSmudge bool `json:"lfsSkipSmudge,omitempty" protobuf:"varint,5,opt,name=lfsSkipSmudge"`
}

// +k8s:openapi-gen=true
type AppFetchHelmChart struct {
	// Example: stable/redis
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// +optional
	Version    string                 `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Repository *AppFetchHelmChartRepo `json:"repository,omitempty" protobuf:"bytes,3,opt,name=repository"`
}

// +k8s:openapi-gen=true
type AppFetchHelmChartRepo struct {
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,2,opt,name=secretRef"`
}

// +k8s:openapi-gen=true
type AppFetchLocalRef struct {
	// Object is expected to be within same namespace
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// +k8s:openapi-gen=true
type AppFetchImgpkgBundle struct {
	Image string `json:"image,omitempty" protobuf:"bytes,1,opt,name=image"`
	// +optional
	TagSelection *versions.VersionSelection `json:"tagSelection,omitempty" protobuf:"bytes,3,opt,name=tagSelection"`
	// Secret may include one or more keys: username, password, token.
	// By default anonymous access is used for authentication.
	// TODO support docker config formated secret
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,2,opt,name=secretRef"`
}
