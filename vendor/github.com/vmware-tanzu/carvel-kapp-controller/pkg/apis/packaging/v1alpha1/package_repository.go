// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pkgr
// +kubebuilder:printcolumn:name=Age,JSONPath=.metadata.creationTimestamp,description=Time since creation,type=date
// +kubebuilder:printcolumn:name=Description,JSONPath=.status.friendlyDescription,description=Friendly description,type=string
type PackageRepository struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PackageRepositorySpec `json:"spec"`
	// +optional
	Status PackageRepositoryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PackageRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PackageRepository `json:"items"`
}

type PackageRepositorySpec struct {
	// Paused when set to true will ignore all pending changes,
	// once it set back to false, pending changes will be applied
	// +optional
	Paused bool `json:"paused,omitempty"`
	// Controls frequency of PackageRepository reconciliation
	// +optional
	SyncPeriod *metav1.Duration `json:"syncPeriod,omitempty"`

	Fetch *PackageRepositoryFetch `json:"fetch"`
}

type PackageRepositoryFetch struct {
	// +optional
	Image *v1alpha1.AppFetchImage `json:"image,omitempty"`
	// +optional
	HTTP *v1alpha1.AppFetchHTTP `json:"http,omitempty"`
	// +optional
	Git *v1alpha1.AppFetchGit `json:"git,omitempty"`
	// +optional
	ImgpkgBundle *v1alpha1.AppFetchImgpkgBundle `json:"imgpkgBundle,omitempty"`
}

type PackageRepositoryStatus struct {
	// +optional
	Fetch *v1alpha1.AppStatusFetch `json:"fetch,omitempty"`
	// +optional
	Template *v1alpha1.AppStatusTemplate `json:"template,omitempty"`
	// +optional
	Deploy *v1alpha1.AppStatusDeploy `json:"deploy,omitempty"`
	// +optional
	ConsecutiveReconcileSuccesses int `json:"consecutiveReconcileSuccesses,omitempty"`
	// +optional
	ConsecutiveReconcileFailures int `json:"consecutiveReconcileFailures,omitempty"`
	// +optional
	v1alpha1.GenericStatus `json:",inline"`
}
