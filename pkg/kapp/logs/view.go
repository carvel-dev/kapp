// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"sync"

	"github.com/cppforlife/go-cli-ui/ui"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type PodWatcher interface {
	Watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) error
}

type View struct {
	tailOpts       PodLogOpts
	podWatcher     PodWatcher
	contFilterFunc func(pod corev1.Pod) []string
	coreClient     kubernetes.Interface
	ui             ui.UI
}

func NewView(
	tailOpts PodLogOpts,
	podWatcher PodWatcher,
	contFilterFunc func(pod corev1.Pod) []string,
	coreClient kubernetes.Interface,
	ui ui.UI,
) View {
	return View{tailOpts, podWatcher, contFilterFunc, coreClient, ui}
}

func (v View) Show(cancelCh chan struct{}) error {
	podsToWatchCh := make(chan corev1.Pod)
	cancelPodTailCh := make(chan struct{})
	cancelPodWatcherCh := make(chan struct{})

	if v.tailOpts.Follow {
		go func() {
			// TODO leaks goroutine
			select {
			case <-cancelCh:
				close(cancelPodWatcherCh)
				close(cancelPodTailCh)
			}
		}()
	} else {
		close(cancelPodWatcherCh)
		// Do not close cancelPodTailCh to let logs stream out on their own
	}

	go func() {
		v.podWatcher.Watch(podsToWatchCh, cancelPodWatcherCh)
		close(podsToWatchCh)
	}()

	var wg sync.WaitGroup

	for pod := range podsToWatchCh {
		pod := pod
		wg.Add(1)

		go func() {
			podsClient := v.coreClient.CoreV1().Pods(pod.Namespace)

			tagFunc := func(cont corev1.Container) string {
				return fmt.Sprintf("%s > %s", pod.Name, cont.Name)
			}

			tailOpts := v.tailOpts
			tailOpts.ContainerNames = v.contFilterFunc(pod)

			err := NewPodLog(pod, podsClient, tagFunc, tailOpts).TailAll(v.ui, cancelPodTailCh)
			if err != nil {
				v.ui.BeginLinef("Pod logs tailing error: %s\n", err)
			}

			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}
