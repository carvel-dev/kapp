// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	appv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	pkgv1alpha1.AddToScheme(scheme.Scheme)
}

type PackagingCarvelDevV1alpha1PackageRepo struct {
	resource ctlres.Resource
}

func NewPackagingCarvelDevV1alpha1PackageRepo(resource ctlres.Resource) *PackagingCarvelDevV1alpha1PackageRepo {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: pkgv1alpha1.SchemeGroupVersion.String(),
		Kind:       "PackageRepository",
	}
	if matcher.Matches(resource) {
		return &PackagingCarvelDevV1alpha1PackageRepo{resource}
	}
	return nil
}

func (s PackagingCarvelDevV1alpha1PackageRepo) IsDoneApplying() DoneApplyState {
	pkgRepo := pkgv1alpha1.PackageRepository{}

	err := s.resource.AsTypedObj(&pkgRepo)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
			"Error: Failed obj conversion: %s", err)}
	}

	if pkgRepo.Generation != pkgRepo.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", pkgRepo.Generation)}
	}

	for _, cond := range pkgRepo.Status.Conditions {
		errorMsg := pkgRepo.Status.UsefulErrorMessage
		if errorMsg == "" {
			errorMsg = cond.Message
		}
		switch {
		case cond.Type == appv1alpha1.Reconciling && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: false, Message: "Reconciling"}

		case cond.Type == appv1alpha1.ReconcileFailed && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
				"Reconcile failed: %s (message: %s)", cond.Reason, errorMsg)}

		case cond.Type == appv1alpha1.DeleteFailed && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
				"Delete failed: %s (message: %s)", cond.Reason, errorMsg)}
		}
	}

	deletingRes := NewDeleting(s.resource)
	if deletingRes != nil {
		return deletingRes.IsDoneApplying()
	}

	return DoneApplyState{Done: true, Successful: true}
}
