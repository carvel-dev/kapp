package resourcesmisc

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
type PodvX struct {
	resource ctlres.Resource
}

func NewPodvX(resource ctlres.Resource) *PodvX {
	matcher := ctlres.APIGroupKindMatcher{
		APIGroup: "",
		Kind:     "Pod",
	}
	if matcher.Matches(resource) {
		return &PodvX{resource}
	}
	return nil
}

func (s PodvX) IsDoneApplying() DoneApplyState {
	// TODO deal with failure scenarios (retry, timeout?)
	if phase, ok := s.resource.Status()["phase"].(string); ok {
		switch phase {
		// Pending: The Pod has been accepted by the Kubernetes system, but one or more of the
		// Container images has not been created. This includes time before being scheduled as
		// well as time spent downloading images over the network, which could take a while.
		case "Pending":
			return DoneApplyState{Done: false, Message: "Pending"}

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
		}
	}

	return DoneApplyState{Done: false, Message: "Undetermined phase"}
}
