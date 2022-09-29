// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // we're unlikely to write descriptive godoc comments in this file.
package v1alpha1

// +k8s:openapi-gen=true
type AppTemplate struct {
	// Use ytt to template configuration
	Ytt *AppTemplateYtt `json:"ytt,omitempty" protobuf:"bytes,1,opt,name=ytt"`
	// Use kbld to resolve image references to use digests
	Kbld *AppTemplateKbld `json:"kbld,omitempty" protobuf:"bytes,2,opt,name=kbld"`
	// Use helm template command to render helm chart
	HelmTemplate *AppTemplateHelmTemplate `json:"helmTemplate,omitempty" protobuf:"bytes,3,opt,name=helmTemplate"`
	Kustomize    *AppTemplateKustomize    `json:"kustomize,omitempty" protobuf:"bytes,4,opt,name=kustomize"`
	Jsonnet      *AppTemplateJsonnet      `json:"jsonnet,omitempty" protobuf:"bytes,5,opt,name=jsonnet"`
	// Use sops to decrypt *.sops.yml files (optional; v0.11.0+)
	Sops *AppTemplateSops `json:"sops,omitempty" protobuf:"bytes,6,opt,name=sops"`
	Cue  *AppTemplateCue  `json:"cue,omitempty" protobuf:"bytes,7,opt,name=cue"`
}

// +k8s:openapi-gen=true
type AppTemplateYtt struct {
	// Ignores comments that ytt doesn't recognize
	// (optional; default=false)
	IgnoreUnknownComments bool `json:"ignoreUnknownComments,omitempty" protobuf:"varint,1,opt,name=ignoreUnknownComments"`
	// Forces strict mode https://github.com/k14s/ytt/blob/develop/docs/strict.md
	// (optional; default=false)
	Strict bool `json:"strict,omitempty" protobuf:"varint,2,opt,name=strict"`
	// Specify additional files, including data values (optional)
	Inline *AppFetchInline `json:"inline,omitempty" protobuf:"bytes,3,opt,name=inline"`
	// Lists paths to provide to ytt explicitly (optional)
	Paths []string `json:"paths,omitempty" protobuf:"bytes,4,rep,name=paths"`
	// Control metadata about input files passed to ytt (optional; v0.18.0+)
	// see https://carvel.dev/ytt/docs/latest/file-marks/ for more details
	FileMarks []string `json:"fileMarks,omitempty" protobuf:"bytes,5,rep,name=fileMarks"`
	// Provide values via ytt's --data-values-file (optional; v0.19.0-alpha.9)
	ValuesFrom []AppTemplateValuesSource `json:"valuesFrom,omitempty" protobuf:"bytes,6,rep,name=valuesFrom"`
}

// +k8s:openapi-gen=true
type AppTemplateKbld struct {
	Paths []string `json:"paths,omitempty" protobuf:"bytes,1,rep,name=paths"`
}

// +k8s:openapi-gen=true
type AppTemplateHelmTemplate struct {
	// Set name explicitly, default is App CR's name (optional; v0.13.0+)
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Set namespace explicitly, default is App CR's namespace (optional; v0.13.0+)
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// Path to chart (optional; v0.13.0+)
	Path string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
	// One or more secrets, config maps, paths that provide values (optional)
	ValuesFrom []AppTemplateValuesSource `json:"valuesFrom,omitempty" protobuf:"bytes,4,rep,name=valuesFrom"`
	// Optional: Get Kubernetes version, defaults (empty) to retrieving the version from the cluster.
	// Can be manually overridden to a value instead.
	KubernetesVersion *Version `json:"kubernetesVersion,omitempty" protobuf:"bytes,5,opt,name=kubernetesVersion"`
	// Optional: Use kubernetes group/versions resources available in the live cluster
	KubernetesAPIs *KubernetesAPIs `json:"kubernetesAPIs,omitempty" protobuf:"bytes,6,opt,name=kubernetesAPIs"`
}

// +k8s:openapi-gen=true
type AppTemplateValuesSource struct {
	SecretRef    *AppTemplateValuesSourceRef   `json:"secretRef,omitempty" protobuf:"bytes,1,opt,name=secretRef"`
	ConfigMapRef *AppTemplateValuesSourceRef   `json:"configMapRef,omitempty" protobuf:"bytes,2,opt,name=configMapRef"`
	Path         string                        `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
	DownwardAPI  *AppTemplateValuesDownwardAPI `json:"downwardAPI,omitempty" protobuf:"bytes,4,opt,name=downwardAPI"`
}

// +k8s:openapi-gen=true
type AppTemplateValuesSourceRef struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// +k8s:openapi-gen=true
type AppTemplateValuesDownwardAPI struct {
	Items []AppTemplateValuesDownwardAPIItem `json:"items,omitempty" protobuf:"bytes,1,opt,name=items"`
}

// +k8s:openapi-gen=true
type AppTemplateValuesDownwardAPIItem struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Required: Selects a field of the app: only annotations, labels, uid, name and namespace are supported.
	FieldPath string `json:"fieldPath,omitempty" protobuf:"bytes,2,opt,name=fieldPath"`
	// Optional: Get running Kubernetes version from cluster, defaults (empty) to retrieving the version from the cluster.
	// Can be manually supplied instead.
	KubernetesVersion *Version `json:"kubernetesVersion,omitempty" protobuf:"bytes,3,opt,name=kubernetesVersion"`
	// Optional: Get running KappController version, defaults (empty) to retrieving the current running version..
	// Can be manually supplied instead.
	KappControllerVersion *Version `json:"kappControllerVersion,omitempty" protobuf:"bytes,4,opt,name=kappControllerVersion"`
	// Optional: Get running KubernetesAPIs from cluster, defaults (empty) to retrieving the APIs from the cluster.
	// Can be manually supplied instead, e.g ["group/version", "group2/version2"]
	KubernetesAPIs *KubernetesAPIs `json:"kubernetesAPIs,omitempty" protobuf:"bytes,5,opt,name=kubernetesAPIs"`
}

// +k8s:openapi-gen=true
type Version struct {
	Version string `json:"version,omitempty" protobuf:"bytes,1,opt,name=version"`
}

// +k8s:openapi-gen=true
type KubernetesAPIs struct {
	GroupVersions []string `json:"groupVersions,omitempty" protobuf:"bytes,1,opt,name=groupVersions"`
}

// TODO implement kustomize
// +k8s:openapi-gen=true
type AppTemplateKustomize struct{}

// TODO implement jsonnet
// +k8s:openapi-gen=true
type AppTemplateJsonnet struct{}

// +k8s:openapi-gen=true
type AppTemplateSops struct {
	// Use PGP to decrypt files (required)
	PGP *AppTemplateSopsPGP `json:"pgp,omitempty" protobuf:"bytes,1,opt,name=pgp"`
	// Lists paths to decrypt explicitly (optional; v0.13.0+)
	Paths []string            `json:"paths,omitempty" protobuf:"bytes,2,rep,name=paths"`
	Age   *AppTemplateSopsAge `json:"age,omitempty" protobuf:"bytes,3,opt,name=age"`
}

// +k8s:openapi-gen=true
type AppTemplateSopsPGP struct {
	// Secret with private armored PGP private keys (required)
	PrivateKeysSecretRef *AppTemplateSopsPrivateKeysSecretRef `json:"privateKeysSecretRef,omitempty" protobuf:"bytes,1,opt,name=privateKeysSecretRef"`
}

// +k8s:openapi-gen=true
type AppTemplateSopsAge struct {
	// Secret with private armored PGP private keys (required)
	PrivateKeysSecretRef *AppTemplateSopsPrivateKeysSecretRef `json:"privateKeysSecretRef,omitempty" protobuf:"bytes,1,opt,name=privateKeysSecretRef"`
}

// +k8s:openapi-gen=true
type AppTemplateSopsPrivateKeysSecretRef struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// +k8s:openapi-gen=true
type AppTemplateCue struct {
	// Explicit list of files/directories (optional)
	Paths []string `json:"paths,omitempty" protobuf:"bytes,1,rep,name=paths"`
	// Provide values (optional)
	ValuesFrom []AppTemplateValuesSource `json:"valuesFrom,omitempty" protobuf:"bytes,2,rep,name=valuesFrom"`
	// Cue expression for single path component, can be used to unify ValuesFrom into a given field (optional)
	InputExpression string `json:"inputExpression,omitempty" protobuf:"bytes,3,opt,name=inputExpression"`
	// Cue expression to output, default will export all visible fields (optional)
	OutputExpression string `json:"outputExpression,omitempty" protobuf:"bytes,4,opt,name=outputExpression"`
}
