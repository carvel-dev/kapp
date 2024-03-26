// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func (r IdentifiedResources) PodResources(labelSelector labels.Selector, resourceNamespaces []string) UniquePodWatcher {
	return UniquePodWatcher{labelSelector, uniqAndValidNamespaces(append(r.fallbackAllowedNamespaces, resourceNamespaces...)), r.coreClient}
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
		namespaces := []string{""}
		namespaces = append(namespaces, w.fallbackAllowedNamespaces...)
		var forbiddenNamespaces []string

		for _, namespace := range namespaces {
			podWatcher := NewPodWatcher(
				w.coreClient.CoreV1().Pods(namespace),
				metav1.ListOptions{LabelSelector: w.labelSelector.String()},
			)
			err := podWatcher.Watch(nonUniquePodsToWatchCh, cancelCh)
			if err == nil {
				if namespace == "" {
					break
				}
				continue
			}
			if !errors.IsForbidden(err) {
				fmt.Printf("Pod watching error: %s\n", err) // TODO
				break
			}
			if namespace != "" {
				forbiddenNamespaces = append(forbiddenNamespaces, fmt.Sprintf(`"%s"`, namespace))
			}
		}

		if len(forbiddenNamespaces) > 0 {
			fmt.Printf(`Pod watching error: pods is forbidden: User cannot list resource "pods" in API group "" in the namespace(s) %s`, strings.Join(forbiddenNamespaces, ", "))
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
