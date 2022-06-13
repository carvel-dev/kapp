// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pkgi,categories={carvel}
// +kubebuilder:printcolumn:name=Package name,JSONPath=.spec.packageRef.refName,description=PackageMetadata name,type=string
// +kubebuilder:printcolumn:name=Package version,JSONPath=.status.version,description=PackageMetadata version,type=string
// +kubebuilder:printcolumn:name=Description,JSONPath=.status.friendlyDescription,description=Friendly description,type=string
// +kubebuilder:printcolumn:name=Age,JSONPath=.metadata.creationTimestamp,description=Time since creation,type=date
// A Package Install is an actual installation of a package and its underlying resources on a Kubernetes cluster.
// It is represented in kapp-controller by a PackageInstall CR.
// A PackageInstall CR must reference a Package CR.
type PackageInstall struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PackageInstallSpec `json:"spec"`
	// +optional
	Status PackageInstallStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PackageInstallList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []PackageInstall `json:"items"`
}

type PackageInstallSpec struct {
	// Specifies service account that will be used to install underlying package contents
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Specifies that Package should be deployed to destination cluster;
	// by default, cluster is same as where this resource resides (optional)
	// +optional
	Cluster *v1alpha1.AppCluster `json:"cluster,omitempty"`
	// Specifies the name of the package to install (required)
	// +optional
	PackageRef *PackageRef `json:"packageRef,omitempty"`
	// Values to be included in package's templating step
	// (currently only included in the first templating step) (optional)
	// +optional
	Values []PackageInstallValues `json:"values,omitempty"`
	// Paused when set to true will ignore all pending changes,
	// once it set back to false, pending changes will be applied
	// +optional
	Paused bool `json:"paused,omitempty"`
	// Canceled when set to true will stop all active changes
	// +optional
	Canceled bool `json:"canceled,omitempty"`
	// Controls frequency of App reconciliation in time + unit
	// format. Always >= 30s. If value below 30s is specified,
	// 30s will be used.
	// +optional
	SyncPeriod *metav1.Duration `json:"syncPeriod,omitempty"`
	// When NoopDelete set to true, PackageInstall deletion
	// should delete PackageInstall/App CR but preserve App's
	// associated resources.
	// +optional
	NoopDelete bool `json:"noopDelete,omitempty"`
}

type PackageRef struct {
	// +optional
	RefName string `json:"refName,omitempty"`
	// +optional
	VersionSelection *versions.VersionSelectionSemver `json:"versionSelection,omitempty"`
}

type PackageInstallValues struct {
	// +optional
	SecretRef *PackageInstallValuesSecretRef `json:"secretRef,omitempty"`
}

type PackageInstallValuesSecretRef struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Key string `json:"key,omitempty"`
}

type PackageInstallStatus struct {
	// +optional
	v1alpha1.GenericStatus `json:",inline"`
	// TODO this is desired resolved version (not actually deployed)
	// +optional
	Version string `json:"version,omitempty"`
	// LastAttemptedVersion specifies what version was last attempted to be installed.
	// It does _not_ indicate it was successfully installed.
	// +optional
	LastAttemptedVersion string `json:"lastAttemptedVersion,omitempty"`
}
