// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
)

// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
type CoreV1Pod struct {
	resource ctlres.Resource
}

func NewCoreV1Pod(resource ctlres.Resource) *CoreV1Pod {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "v1",
		Kind:       "Pod",
	}
	if matcher.Matches(resource) {
		return &CoreV1Pod{resource}
	}
	return nil
}

func (s CoreV1Pod) IsDoneApplying() DoneApplyState {
	pod := corev1.Pod{}

	err := s.resource.AsTypedObj(&pod)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf("Error: Failed obj conversion: %s", err)}
	}

	// TODO deal with failure scenarios (retry, timeout?)
	switch pod.Status.Phase {
	// Pending: The Pod has been accepted by the Kubernetes system, but one or more of the
	// Container images has not been created. This includes time before being scheduled as
	// well as time spent downloading images over the network, which could take a while.
	case "Pending":
		return DoneApplyState{Done: false, Message: s.detailedMsg("Pending", s.pendingDetailsReason(pod))}

	// Running: The Pod has been bound to a node, and all of the Containers have been created.
	// At least one Container is still running, or is in the process of starting or restarting.
	case "Running":
		allTrue, msg := Conditions{s.resource}.IsSelectedTrue([]string{"Initialized", "Ready", "PodScheduled"})
		return DoneApplyState{Done: allTrue, Successful: allTrue, Message: msg}

	// Succeeded: All Containers in the Pod have terminated in success, and will not be restarted.
	case "Succeeded":
		return DoneApplyState{Done: true, Successful: true, Message: ""}

	// Failed: All Containers in the Pod have terminated, and at least one Container has
	// terminated in failure. That is, the Container either exited with non-zero status
	// or was terminated by the system.
	case "Failed":
		return DoneApplyState{Done: true, Successful: false, Message: "Phase is failed"}

	// Unknown: For some reason the state of the Pod could not be obtained,
	// typically due to an error in communicating with the host of the Pod.
	case "Unknown":
		return DoneApplyState{Done: true, Successful: false, Message: "Phase is unknown"}

	default:
		return DoneApplyState{Done: false, Message: "Undetermined phase"}
	}
}

func (s CoreV1Pod) detailedMsg(state, msg string) string {
	if len(msg) > 0 {
		return state + ": " + msg
	}
	return state
}

func (s CoreV1Pod) pendingDetailsReason(pod corev1.Pod) string {
	statuses := append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	statuses = append(statuses, pod.Status.ContainerStatuses...)

	// See status example below
	for _, st := range statuses {
		if st.State.Waiting != nil {
			msg := st.State.Waiting.Reason
			if len(st.State.Waiting.Message) > 0 {
				msg += fmt.Sprintf(" (message: %s)", st.State.Waiting.Message)
			}
			return msg
		}
	}

	// See status example below
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodScheduled {
			if cond.Reason == corev1.PodReasonUnschedulable {
				msg := cond.Reason
				if len(cond.Message) > 0 {
					msg += fmt.Sprintf(" (message: %s)", cond.Message)
				}
				return msg
			}
		}
	}

	return ""
}

/*

# Image cannot be pulled
status:
  containerStatuses:
  - image: kbld:docker-io-...
    imageID: ""
    lastState: {}
    name: simple-app
    ready: false
    restartCount: 0
    state:
      waiting:
        message: 'rpc error: code = Unknown desc = Error response from daemon: repository
          kbld not found: does not exist or no pull access'
        reason: ErrImagePull

# Unschedulable pod
status:
  conditions:
  - lastProbeTime: "2019-07-16T23:53:29Z"
    lastTransitionTime: "2019-07-16T23:51:54Z"
    message: '0/4 nodes are available: 3 node(s) didn''t match node selector, 4 node(s)
      didn''t have free ports for the requested pod ports.'
    reason: Unschedulable
    status: "False"
    type: PodScheduled
  phase: Pending
  qosClass: Burstable

*/
