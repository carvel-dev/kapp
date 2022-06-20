// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
)

// +k8s:openapi-gen=true
type AppFetch struct {
	// Pulls content from within this resource; or other resources in the cluster
	Inline *AppFetchInline `json:"inline,omitempty" protobuf:"bytes,1,opt,name=inline"`
	// Pulls content from Docker/OCI registry
	Image *AppFetchImage `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	// Uses http library to fetch file
	HTTP *AppFetchHTTP `json:"http,omitempty" protobuf:"bytes,3,opt,name=http"`
	// Uses git to clone repository
	Git *AppFetchGit `json:"git,omitempty" protobuf:"bytes,4,opt,name=git"`
	// Uses helm fetch to fetch specified chart
	HelmChart *AppFetchHelmChart `json:"helmChart,omitempty" protobuf:"bytes,5,opt,name=helmChart"`
	// Pulls imgpkg bundle from Docker/OCI registry (v0.17.0+)
	ImgpkgBundle *AppFetchImgpkgBundle `json:"imgpkgBundle,omitempty" protobuf:"bytes,6,opt,name=imgpkgBundle"`
	// Relative path to place the fetched artifacts
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,7,opt,name=path"`
}

// +k8s:openapi-gen=true
type AppFetchInline struct {
	// Specifies mapping of paths to their content;
	// not recommended for sensitive values as CR is not encrypted (optional)
	Paths map[string]string `json:"paths,omitempty" protobuf:"bytes,1,rep,name=paths"`
	// Specifies content via secrets and config maps;
	// data values are recommended to be placed in secrets (optional)
	PathsFrom []AppFetchInlineSource `json:"pathsFrom,omitempty" protobuf:"bytes,2,rep,name=pathsFrom"`
}

// +k8s:openapi-gen=true
type AppFetchInlineSource struct {
	SecretRef    *AppFetchInlineSourceRef `json:"secretRef,omitempty" protobuf:"bytes,1,opt,name=secretRef"`
	ConfigMapRef *AppFetchInlineSourceRef `json:"configMapRef,omitempty" protobuf:"bytes,2,opt,name=configMapRef"`
}

// +k8s:openapi-gen=true
type AppFetchInlineSourceRef struct {
	// Specifies where to place files found in secret (optional)
	DirectoryPath string `json:"directoryPath,omitempty" protobuf:"bytes,2,opt,name=directoryPath"`
	Name          string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// +k8s:openapi-gen=true
type AppFetchImage struct {
	// Docker image url; unqualified, tagged, or
	// digest references supported (required)
	// Example: username/app1-config:v0.1.0
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// Specifies a strategy to choose a tag (optional; v0.24.0+)
	// if specified, do not include a tag in url key
	// +optional
	TagSelection *versions.VersionSelection `json:"tagSelection,omitempty" protobuf:"bytes,4,opt,name=tagSelection"`
	// Secret may include one or more keys: username, password, token.
	// By default anonymous access is used for authentication.
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,2,opt,name=secretRef"`
	// Grab only portion of image (optional)
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,3,opt,name=subPath"`
}

// +k8s:openapi-gen=true
type AppFetchHTTP struct {
	// URL can point to one of following formats: text, tgz, zip
	// http and https url are supported;
	// plain file, tgz and tar types are supported (required)
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// Checksum to verify after download (optional)
	// +optional
	SHA256 string `json:"sha256,omitempty" protobuf:"bytes,2,opt,name=sha256"`
	// Secret to provide auth details (optional)
	// Secret may include one or more keys: username, password
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,3,opt,name=secretRef"`
	// Grab only portion of download (optional)
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`
}

// +k8s:openapi-gen=true
type AppFetchGit struct {
	// http or ssh urls are supported (required)
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// Branch, tag, commit; origin is the name of the remote (optional)
	// +optional
	Ref string `json:"ref,omitempty" protobuf:"bytes,2,opt,name=ref"`
	// Specifies a strategy to resolve to an explicit ref (optional; v0.24.0+)
	// +optional
	RefSelection *versions.VersionSelection `json:"refSelection,omitempty" protobuf:"bytes,6,opt,name=refSelection"`
	// Secret with auth details. allowed keys: ssh-privatekey, ssh-knownhosts, username, password (optional)
	// (if ssh-knownhosts is not specified, git will not perform strict host checking)
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,3,opt,name=secretRef"`
	// Grab only portion of repository (optional)
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`
	// Skip lfs download (optional)
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
	// Repository url;
	// scheme of oci:// will fetch experimental helm oci chart (v0.19.0+)
	// (required)
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
	// Docker image url; unqualified, tagged, or
	// digest references supported (required)
	Image string `json:"image,omitempty" protobuf:"bytes,1,opt,name=image"`
	// Specifies a strategy to choose a tag (optional; v0.24.0+)
	// if specified, do not include a tag in url key
	// +optional
	TagSelection *versions.VersionSelection `json:"tagSelection,omitempty" protobuf:"bytes,3,opt,name=tagSelection"`
	// Secret may include one or more keys: username, password, token.
	// By default anonymous access is used for authentication.
	// +optional
	SecretRef *AppFetchLocalRef `json:"secretRef,omitempty" protobuf:"bytes,2,opt,name=secretRef"`
}
