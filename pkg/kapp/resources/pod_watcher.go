// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodWatcher struct {
	podsClient typedcorev1.PodInterface
	listOpts   metav1.ListOptions
}

func NewPodWatcher(
	podsClient typedcorev1.PodInterface,
	listOpts metav1.ListOptions,
) PodWatcher {
	return PodWatcher{podsClient, listOpts}
}

func (w PodWatcher) Watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) error {
	podsList, err := w.podsClient.List(context.TODO(), w.listOpts)
	if err != nil {
		return err
	}

	for _, pod := range podsList.Items {
		podsToWatchCh <- pod
	}

	// Return before potentially getting any events
	select {
	case <-cancelCh:
		return nil
	default:
	}

	for {
		retry, err := w.watch(podsToWatchCh, cancelCh)
		if err != nil {
			return err
		}
		if !retry {
			return nil
		}
	}
}

func (w PodWatcher) watch(podsToWatchCh chan corev1.Pod, cancelCh chan struct{}) (bool, error) {
	watcher, err := w.podsClient.Watch(context.TODO(), w.listOpts)
	if err != nil {
		return false, fmt.Errorf("Creating Pod watcher: %s", err)
	}

	defer watcher.Stop()

	for {
		select {
		case e, ok := <-watcher.ResultChan():
			if !ok || e.Object == nil {
				// Watcher may expire, hence try to retry
				return true, nil
			}

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			switch e.Type {
			case watch.Added:
				podsToWatchCh <- *pod
			}

		case <-cancelCh:
			return false, nil
		}
	}
}
