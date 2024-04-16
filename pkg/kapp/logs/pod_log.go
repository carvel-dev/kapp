// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"sync"

	"carvel.dev/kapp/pkg/kapp/matcher"
	"github.com/cppforlife/go-cli-ui/ui"
	corev1 "k8s.io/api/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodLogOpts struct {
	Follow         bool
	Lines          *int64
	ContainerNames []string
	ContainerTag   bool
	LinePrefix     string
}

type PodLog struct {
	pod        corev1.Pod
	podsClient typedcorev1.PodInterface

	tagFunc func(corev1.Container) string
	opts    PodLogOpts
}

func NewPodLog(
	pod corev1.Pod,
	podsClient typedcorev1.PodInterface,
	tagFunc func(corev1.Container) string,
	opts PodLogOpts,
) PodLog {
	return PodLog{pod, podsClient, tagFunc, opts}
}

// TailAll will tail all logs from all containers in a single Pod
func (l PodLog) TailAll(ui ui.UI, cancelCh chan struct{}) error {
	// Container will not emit any new logs since this is a terminal position
	podInTerminalState := l.pod.Status.Phase == corev1.PodSucceeded || l.pod.Status.Phase == corev1.PodFailed

	var conts []corev1.Container

	for _, cont := range l.pod.Spec.InitContainers {
		if !(podInTerminalState && l.isWaitingContainer(cont, l.pod.Status.InitContainerStatuses)) {
			if l.isWatchingContainer(cont, l.opts.ContainerNames) {
				conts = append(conts, cont)
			}
		}
	}

	for _, cont := range l.pod.Spec.Containers {
		if !(podInTerminalState && l.isWaitingContainer(cont, l.pod.Status.ContainerStatuses)) {
			if l.isWatchingContainer(cont, l.opts.ContainerNames) {
				conts = append(conts, cont)
			}
		}
	}

	var wg sync.WaitGroup

	for _, cont := range conts {
		cont := cont
		wg.Add(1)

		go func() {
			NewPodContainerLog(l.pod, cont.Name, l.podsClient, l.tagFunc(cont), l.opts).Tail(ui, cancelCh) // TODO err?
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}

func (l PodLog) isWaitingContainer(cont corev1.Container, statuses []corev1.ContainerStatus) bool {
	for _, contStatus := range statuses {
		if cont.Name == contStatus.Name {
			return contStatus.State.Waiting != nil
		}
	}
	return false
}

func (l PodLog) isWatchingContainer(cont corev1.Container, containers []string) bool {
	if len(containers) == 0 {
		return true
	}
	for _, n := range containers {
		if matcher.NewStringMatcher(n).Matches(cont.Name) {
			return true
		}
	}
	return false
}
