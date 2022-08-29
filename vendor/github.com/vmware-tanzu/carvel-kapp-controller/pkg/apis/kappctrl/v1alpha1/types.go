// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={carvel}
// +kubebuilder:printcolumn:name=Description,JSONPath=.status.friendlyDescription,description=Friendly description,type=string
// +kubebuilder:printcolumn:name=Since-Deploy,JSONPath=.status.deploy.startedAt,description=Last time app started being deployed. Does not mean anything was changed.,type=date
// +kubebuilder:printcolumn:name=Age,JSONPath=.metadata.creationTimestamp,description=Time since creation,type=date
// +protobuf=false
// An App is a set of Kubernetes resources. These resources could span any number of namespaces or could be cluster-wide (e.g. CRDs). An App is represented in kapp-controller using a App CR.
// The App CR comprises of three main sections:
// spec.fetch – declare source for fetching configuration and OCI images
// spec.template – declare templating tool and values
// spec.deploy – declare deployment tool and any deploy specific configuration
type App struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AppSpec `json:"spec"`
	// +optional
	Status AppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +protobuf=false
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []App `json:"items"`
}

// +k8s:openapi-gen=true
type AppSpec struct {
	// Specifies that app should be deployed authenticated via
	// given service account, found in this namespace (optional; v0.6.0+)
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,1,opt,name=serviceAccountName"`
	// Specifies that app should be deployed to destination cluster;
	// by default, cluster is same as where this resource resides (optional; v0.5.0+)
	// +optional
	Cluster *AppCluster `json:"cluster,omitempty" protobuf:"bytes,2,opt,name=cluster"`
	// +optional
	Fetch []AppFetch `json:"fetch,omitempty" protobuf:"bytes,3,rep,name=fetch"`
	// +optional
	Template []AppTemplate `json:"template,omitempty" protobuf:"bytes,4,rep,name=template"`
	// +optional
	Deploy []AppDeploy `json:"deploy,omitempty" protobuf:"bytes,5,rep,name=deploy"`
	// Pauses _future_ reconciliation; does _not_ affect
	// currently running reconciliation (optional; default=false)
	// +optional
	Paused bool `json:"paused,omitempty" protobuf:"varint,6,opt,name=paused"`
	// Cancels current and future reconciliations (optional; default=false)
	// +optional
	Canceled bool `json:"canceled,omitempty" protobuf:"varint,7,opt,name=canceled"`
	// Specifies the length of time to wait, in time + unit
	// format, before reconciling. Always >= 30s. If value below
	// 30s is specified, 30s will be used. (optional; v0.9.0+; default=30s)
	// +optional
	SyncPeriod *metav1.Duration `json:"syncPeriod,omitempty" protobuf:"bytes,8,opt,name=syncPeriod"`
	// Deletion requests for the App will result in the App CR being
	// deleted, but its associated resources will not be deleted
	// (optional; default=false; v0.18.0+)
	// +optional
	NoopDelete bool `json:"noopDelete,omitempty" protobuf:"varint,9,opt,name=noopDelete"`
}

// +k8s:openapi-gen=true
type AppCluster struct {
	// Specifies namespace in destination cluster (optional)
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Specifies secret containing kubeconfig (required)
	// +optional
	KubeconfigSecretRef *AppClusterKubeconfigSecretRef `json:"kubeconfigSecretRef,omitempty" protobuf:"bytes,2,opt,name=kubeconfigSecretRef"`
}

// +k8s:openapi-gen=true
type AppClusterKubeconfigSecretRef struct {
	// Specifies secret name within app's namespace (required)
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Specifies key that contains kubeconfig (optional)
	// +optional
	Key string `json:"key,omitempty" protobuf:"bytes,2,opt,name=key"`
}

// +protobuf=false
type AppStatus struct {
	// +optional
	ManagedAppName string `json:"managedAppName,omitempty"`
	// +optional
	Fetch *AppStatusFetch `json:"fetch,omitempty"`
	// +optional
	Template *AppStatusTemplate `json:"template,omitempty"`
	// +optional
	Deploy *AppStatusDeploy `json:"deploy,omitempty"`
	// +optional
	Inspect *AppStatusInspect `json:"inspect,omitempty"`
	// +optional
	ConsecutiveReconcileSuccesses int `json:"consecutiveReconcileSuccesses,omitempty"`
	// +optional
	ConsecutiveReconcileFailures int `json:"consecutiveReconcileFailures,omitempty"`
	// +optional
	GenericStatus `json:",inline"`
}

// +protobuf=false
type AppStatusFetch struct {
	// +optional
	Stderr string `json:"stderr,omitempty"`
	// +optional
	Stdout string `json:"stdout,omitempty"`
	// +optional
	ExitCode int `json:"exitCode"`
	// +optional
	Error string `json:"error,omitempty"`
	// +optional
	StartedAt metav1.Time `json:"startedAt,omitempty"`
	// +optional
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
}

// +protobuf=false
type AppStatusTemplate struct {
	// +optional
	Stderr string `json:"stderr,omitempty"`
	// +optional
	ExitCode int `json:"exitCode"`
	// +optional
	Error string `json:"error,omitempty"`
	// +optional
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
}

// +protobuf=false
type AppStatusDeploy struct {
	// +optional
	Stdout string `json:"stdout,omitempty"`
	// +optional
	Stderr string `json:"stderr,omitempty"`
	// +optional
	Finished bool `json:"finished"`
	// +optional
	ExitCode int `json:"exitCode"`
	// +optional
	Error string `json:"error,omitempty"`
	// +optional
	StartedAt metav1.Time `json:"startedAt,omitempty"`
	// +optional
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
	// +optional
	KappDeployStatus *KappDeployStatus `json:"kapp,omitempty"`
}

// KappDeployStatus contains the associated AppCR deployed resources
// +protobuf=false
type KappDeployStatus struct {
	AssociatedResources AssociatedResources `json:"associatedResources,omitempty"`
}

// AssociatedResources contains the associated App label, namespaces and GKs
// +protobuf=false
type AssociatedResources struct {
	Label      string             `json:"label,omitempty"`
	Namespaces []string           `json:"namespaces,omitempty"`
	GroupKinds []metav1.GroupKind `json:"groupKinds,omitempty"`
}

// +protobuf=false
type AppStatusInspect struct {
	// +optional
	Stdout string `json:"stdout,omitempty"`
	// +optional
	Stderr string `json:"stderr,omitempty"`
	// +optional
	ExitCode int `json:"exitCode"`
	// +optional
	Error string `json:"error,omitempty"`
	// +optional
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
}
