// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type GenericStatus struct {
	// Populated based on metadata.generation when controller
	// observes a change to the resource; if this value is
	// out of data, other status fields do not reflect latest state
	// +optional
	ObservedGeneration int64 `json:"observedGeneration" protobuf:"varint,1,opt,name=observedGeneration"`
	// +optional
	Conditions []Condition `json:"conditions" protobuf:"bytes,2,rep,name=conditions"`
	// +optional
	FriendlyDescription string `json:"friendlyDescription" protobuf:"bytes,3,opt,name=friendlyDescription"`
	// +optional
	UsefulErrorMessage string `json:"usefulErrorMessage,omitempty" protobuf:"bytes,4,opt,name=usefulErrorMessage"`
}

// ConditionType represents reconciler state
type ConditionType string

// Constants representing reconciler state as a condition
const (
	Reconciling        ConditionType = "Reconciling"
	ReconcileFailed    ConditionType = "ReconcileFailed"
	ReconcileSucceeded ConditionType = "ReconcileSucceeded"

	Deleting     ConditionType = "Deleting"
	DeleteFailed ConditionType = "DeleteFailed"
)

// +k8s:openapi-gen=true
type Condition struct {
	Type   ConditionType          `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ConditionType"`
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Unique, this should be a short, machine understandable string that gives the reason
	// for condition's last transition. If it reports "ResizeStarted" that means the underlying
	// persistent volume is being resized.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
}
