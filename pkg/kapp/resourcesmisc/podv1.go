package resourcesmisc

import (
	"fmt"

	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
)

// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
type Podv1 struct {
	resource ctlres.Resource
}

func NewPodv1(resource ctlres.Resource) *Podv1 {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: "v1",
		Kind:       "Pod",
	}
	if matcher.Matches(resource) {
		return &Podv1{resource}
	}
	return nil
}

func (s Podv1) IsDoneApplying() DoneApplyState {
	// TODO deal with failure scenarios (retry, timeout?)
	if phase, ok := s.resource.Status()["phase"].(string); ok {
		switch phase {
		// Pending: The Pod has been accepted by the Kubernetes system, but one or more of the
		// Container images has not been created. This includes time before being scheduled as
		// well as time spent downloading images over the network, which could take a while.
		case "Pending":
			return DoneApplyState{Done: false, Message: s.detailedMsg("Pending", s.pendingDetailsReason())}

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

func (s Podv1) detailedMsg(state, msg string) string {
	if len(msg) > 0 {
		return state + ": " + msg
	}
	return state
}

func (s Podv1) pendingDetailsReason() string {
	pod := corev1.Pod{}

	err := s.resource.AsTypedObj(&pod)
	if err != nil {
		return ""
	}

	statuses := append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	statuses = append(statuses, pod.Status.ContainerStatuses...)

	for _, st := range statuses {
		if st.State.Waiting != nil {
			return fmt.Sprintf("%s (message: %s)",
				st.State.Waiting.Reason, st.State.Waiting.Message)
		}
	}

	return ""
}

/*

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

*/
