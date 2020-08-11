// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func (r IdentifiedResources) PodResources(labelSelector labels.Selector) UniquePodWatcher {
	return UniquePodWatcher{labelSelector, r.fallbackAllowedNamespaces, r.coreClient}
}

type PodWatcherI interface {
	Watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) error
}

type UniquePodWatcher struct {
	labelSelector             labels.Selector
	fallbackAllowedNamespaces []string
	coreClient                kubernetes.Interface
}

var _ PodWatcherI = UniquePodWatcher{}

func (w UniquePodWatcher) Watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) error {
	nonUniquePodsToWatchCh := make(chan corev1.Pod)

	go func() {
		// Watch Pods in all namespaces first and fallback to the
		// fallbackAllowedNamespaces if lack of permission
		namespace := ""
		for {
			podWatcher := NewPodWatcher(
				w.coreClient.CoreV1().Pods(namespace),
				metav1.ListOptions{LabelSelector: w.labelSelector.String()},
			)

			err := podWatcher.Watch(nonUniquePodsToWatchCh, cancelCh)
			if err == nil {
				break
			}
			if errors.IsForbidden(err) && namespace == "" {
				// The '-n' flag or default state namespace can specify only 1 namespace, so there
				// should be at most 1 item in fallbackAllowedNamespaces
				if len(w.fallbackAllowedNamespaces) > 0 {
					namespace = w.fallbackAllowedNamespaces[0]
					if namespace == "" {
						break
					}
				}
			} else {
				fmt.Printf("Pod watching error: %s\n", err) // TODO
				break
			}
		}

		close(nonUniquePodsToWatchCh)
	}()

	// Send unique pods to the watcher client
	watchedPods := map[string]struct{}{}

	for pod := range nonUniquePodsToWatchCh {
		podUID := string(pod.UID)
		if _, found := watchedPods[podUID]; found {
			continue
		}

		watchedPods[podUID] = struct{}{}
		podsToWatchCh <- pod
	}

	return nil
}

type FilteringPodWatcher struct {
	MatcherFunc func(*corev1.Pod) bool
	Watcher     PodWatcherI
}

var _ PodWatcherI = FilteringPodWatcher{}

func (w FilteringPodWatcher) Watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) error {
	filteredCh := make(chan corev1.Pod)

	go func() {
		err := w.Watcher.Watch(filteredCh, cancelCh)
		if err != nil {
			fmt.Printf("Pod watching error: %s\n", err) // TODO
		}

		close(filteredCh)
	}()

	for pod := range filteredCh {
		if w.MatcherFunc(&pod) {
			podsToWatchCh <- pod
		}
	}

	return nil
}
