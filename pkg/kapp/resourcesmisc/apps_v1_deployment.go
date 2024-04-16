// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"
	"strconv"
	"strings"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	appsV1DeploymentWaitMinimumReplicasAvailableAnnKey = "kapp.k14s.io/apps-v1-deployment-wait-minimum-replicas-available" // values: "10", "5%"
)

type AppsV1Deployment struct {
	resource     ctlres.Resource
	associatedRs []ctlres.Resource
}

func NewAppsV1Deployment(resource ctlres.Resource, associatedRs []ctlres.Resource) *AppsV1Deployment {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}
	if matcher.Matches(resource) {
		return &AppsV1Deployment{resource, associatedRs}
	}
	return nil
}

func (s AppsV1Deployment) IsDoneApplying() DoneApplyState {
	dep := appsv1.Deployment{}

	err := s.resource.AsTypedObj(&dep)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	if dep.Generation != dep.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", dep.Generation)}
	}

	for _, cond := range dep.Status.Conditions {
		switch cond.Type {
		case appsv1.DeploymentProgressing:
			if cond.Status == corev1.ConditionFalse {
				return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
					"Deployment is not progressing: %s (message: %s)", cond.Reason, cond.Message)}
			}

		case appsv1.DeploymentReplicaFailure:
			if cond.Status == corev1.ConditionTrue {
				return DoneApplyState{Done: false, Successful: false, Message: fmt.Sprintf(
					"Deployment has encountered replica failure: %s (message: %s)", cond.Reason, cond.Message)}
			}
		}
	}

	// TODO ideally we would not condition this on len of associated resources
	if len(s.associatedRs) > 0 {
		minRepAvailable, found := s.resource.Annotations()[appsV1DeploymentWaitMinimumReplicasAvailableAnnKey]
		if found {
			return s.isMinReplicasAvailable(dep, minRepAvailable)
		}
	}

	if dep.Status.UnavailableReplicas > 0 {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d unavailable replicas", dep.Status.UnavailableReplicas)}
	}

	if dep.Spec.Replicas != nil && dep.Status.ReadyReplicas < *dep.Spec.Replicas {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for %d/%d replicas to be ready", *dep.Spec.Replicas-dep.Status.ReadyReplicas, *dep.Spec.Replicas)}
	}
	return DoneApplyState{Done: true, Successful: true}
}

func (s AppsV1Deployment) isMinReplicasAvailable(dep appsv1.Deployment, expectedMinRepAvailableStr string) DoneApplyState {
	isPercent := strings.HasSuffix(expectedMinRepAvailableStr, "%")

	minRepAvailable, err := strconv.Atoi(strings.TrimSuffix(expectedMinRepAvailableStr, "%"))
	if err != nil {
		return DoneApplyState{Done: true, Successful: false,
			Message: fmt.Sprintf("Error: Failed to parse %s: %s", appsV1DeploymentWaitMinimumReplicasAvailableAnnKey, err)}
	}

	if dep.Spec.Replicas == nil {
		return DoneApplyState{Done: true, Successful: false,
			Message: fmt.Sprintf("Error: Failed to find spec.replicas")}
	}

	totalReplicas := int(*dep.Spec.Replicas)

	if isPercent {
		minRepAvailable = totalReplicas * minRepAvailable / 100
	}

	if minRepAvailable > totalReplicas {
		minRepAvailable = totalReplicas
	}
	if totalReplicas > 0 && minRepAvailable <= 0 {
		minRepAvailable = 1
	}

	rs, err := s.findLatestReplicaSet(dep)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false,
			Message: fmt.Sprintf("Error: Failed to find latest replicaset: %s", err)}
	}

	return rs.IsDoneApplyingWithMinimum(minRepAvailable)
}

const (
	deploymentRevAnnKey = "deployment.kubernetes.io/revision"
)

func (s AppsV1Deployment) findLatestReplicaSet(dep appsv1.Deployment) (*ExtensionsAndAppsVxReplicaSet, error) {
	expectedRevKey, found := dep.Annotations[deploymentRevAnnKey]
	if !found {
		return nil, fmt.Errorf("Expected to find '%s' but did not", deploymentRevAnnKey)
	}

	for _, res := range s.associatedRs {
		// Cannot use appsv1 RS since no gurantee which versions are in associated resources
		rs := NewExtensionsAndAppsVxReplicaSet(res)
		if rs != nil && res.Annotations()[deploymentRevAnnKey] == expectedRevKey {
			return rs, nil
		}
	}

	return nil, fmt.Errorf("Expected to find replica set (rev %s) in associated resources but did not", expectedRevKey)
}
