// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// +k8s:openapi-gen=true
type AppTemplate struct {
	Ytt          *AppTemplateYtt          `json:"ytt,omitempty" protobuf:"bytes,1,opt,name=ytt"`
	Kbld         *AppTemplateKbld         `json:"kbld,omitempty" protobuf:"bytes,2,opt,name=kbld"`
	HelmTemplate *AppTemplateHelmTemplate `json:"helmTemplate,omitempty" protobuf:"bytes,3,opt,name=helmTemplate"`
	Kustomize    *AppTemplateKustomize    `json:"kustomize,omitempty" protobuf:"bytes,4,opt,name=kustomize"`
	Jsonnet      *AppTemplateJsonnet      `json:"jsonnet,omitempty" protobuf:"bytes,5,opt,name=jsonnet"`
	Sops         *AppTemplateSops         `json:"sops,omitempty" protobuf:"bytes,6,opt,name=sops"`
}

// +k8s:openapi-gen=true
type AppTemplateYtt struct {
	IgnoreUnknownComments bool                      `json:"ignoreUnknownComments,omitempty" protobuf:"varint,1,opt,name=ignoreUnknownComments"`
	Strict                bool                      `json:"strict,omitempty" protobuf:"varint,2,opt,name=strict"`
	Inline                *AppFetchInline           `json:"inline,omitempty" protobuf:"bytes,3,opt,name=inline"`
	Paths                 []string                  `json:"paths,omitempty" protobuf:"bytes,4,rep,name=paths"`
	FileMarks             []string                  `json:"fileMarks,omitempty" protobuf:"bytes,5,rep,name=fileMarks"`
	ValuesFrom            []AppTemplateValuesSource `json:"valuesFrom,omitempty" protobuf:"bytes,6,rep,name=valuesFrom"`
}

// +k8s:openapi-gen=true
type AppTemplateKbld struct {
	Paths []string `json:"paths,omitempty" protobuf:"bytes,1,rep,name=paths"`
}

// +k8s:openapi-gen=true
type AppTemplateHelmTemplate struct {
	Name       string                    `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Namespace  string                    `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	Path       string                    `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
	ValuesFrom []AppTemplateValuesSource `json:"valuesFrom,omitempty" protobuf:"bytes,4,rep,name=valuesFrom"`
}

// +k8s:openapi-gen=true
type AppTemplateValuesSource struct {
	SecretRef    *AppTemplateValuesSourceRef `json:"secretRef,omitempty" protobuf:"bytes,1,opt,name=secretRef"`
	ConfigMapRef *AppTemplateValuesSourceRef `json:"configMapRef,omitempty" protobuf:"bytes,2,opt,name=configMapRef"`
	Path         string                      `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`
}

// +k8s:openapi-gen=true
type AppTemplateValuesSourceRef struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// TODO implement kustomize
// +k8s:openapi-gen=true
type AppTemplateKustomize struct{}

// TODO implement jsonnet
// +k8s:openapi-gen=true
type AppTemplateJsonnet struct{}

// +k8s:openapi-gen=true
type AppTemplateSops struct {
	PGP   *AppTemplateSopsPGP `json:"pgp,omitempty" protobuf:"bytes,1,opt,name=pgp"`
	Paths []string            `json:"paths,omitempty" protobuf:"bytes,2,rep,name=paths"`
}

// +k8s:openapi-gen=true
type AppTemplateSopsPGP struct {
	PrivateKeysSecretRef *AppTemplateSopsPGPPrivateKeysSecretRef `json:"privateKeysSecretRef,omitempty" protobuf:"bytes,1,opt,name=privateKeysSecretRef"`
}

// +k8s:openapi-gen=true
type AppTemplateSopsPGPPrivateKeysSecretRef struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}
