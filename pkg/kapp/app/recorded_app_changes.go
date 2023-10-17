// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

const (
	isChangeLabelKey     = "kapp.k14s.io/is-app-change"
	isChangeLabelValue   = ""
	legacyChangeLabelKey = "kapp.k14s.io/app-change-app"       // holds app name
	changeLabelKey       = "kapp.k14s.io/app-change-app-label" // holds app label
)

type RecordedAppChanges struct {
	nsName           string
	appName          string
	changeLabelValue string

	appChangeUsesAppLabel bool

	coreClient kubernetes.Interface
}

func NewRecordedAppChanges(nsName, appName, changeLabelValue string, appChangeUsesAppLabel bool, coreClient kubernetes.Interface) RecordedAppChanges {
	return RecordedAppChanges{nsName, appName, changeLabelValue, appChangeUsesAppLabel, coreClient}
}

func (a RecordedAppChanges) List() ([]Change, error) {
	var result []Change

	listOpts := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			isChangeLabelKey: isChangeLabelValue,
			changeLabelKey:   a.changeLabelValue,
		}).String(),
	}

	if !a.appChangeUsesAppLabel {
		listOpts = metav1.ListOptions{
			LabelSelector: labels.Set(map[string]string{
				isChangeLabelKey:     isChangeLabelValue,
				legacyChangeLabelKey: a.appName,
			}).String(),
		}
	}

	changes, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}

	sort.Slice(changes.Items, func(i, j int) bool {
		iT := &changes.Items[i].CreationTimestamp
		jT := &changes.Items[j].CreationTimestamp
		return iT.Before(jT)
	})

	for _, change := range changes.Items {
		result = append(result, &ChangeImpl{
			name:       change.Name,
			nsName:     a.nsName,
			coreClient: a.coreClient,
			meta:       NewChangeMetaFromData(change.Data),
			createdAt:  change.CreationTimestamp.Time,
		})
	}

	return result, nil
}

func (a RecordedAppChanges) DeleteAll() error {
	listOpts := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			isChangeLabelKey: isChangeLabelValue,
			changeLabelKey:   a.changeLabelValue,
		}).String(),
	}

	if !a.appChangeUsesAppLabel {
		listOpts = metav1.ListOptions{
			LabelSelector: labels.Set(map[string]string{
				isChangeLabelKey:     isChangeLabelValue,
				legacyChangeLabelKey: a.appName,
			}).String(),
		}
	}

	changes, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).List(context.TODO(), listOpts)
	if err != nil {
		return err
	}

	for _, change := range changes.Items {
		err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Delete(context.TODO(), change.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a RecordedAppChanges) Begin(meta ChangeMeta) (*ChangeImpl, error) {
	newMeta := ChangeMeta{
		StartedAt:   time.Now().UTC(),
		Description: meta.Description,
		Namespaces:  meta.Namespaces,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: a.appName + "-change-",
			Namespace:    a.nsName,
			Labels: map[string]string{
				isChangeLabelKey: isChangeLabelValue,
				changeLabelKey:   a.changeLabelValue,
			},
		},
		Data: newMeta.AsData(),
	}

	// Keep app changes backward compatible if possible, by adding legacy
	// change label key when app name's length is less than maximum allowed length of a label
	if !a.appChangeUsesAppLabel || len(a.appName) <= validation.LabelValueMaxLength {
		configMap.ObjectMeta.Labels[legacyChangeLabelKey] = a.appName
	}

	createdChange, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("Creating app change: %w", err)
	}

	change := &ChangeImpl{
		name:       createdChange.Name,
		nsName:     createdChange.Namespace,
		coreClient: a.coreClient,
		meta:       newMeta,
	}

	return change, nil
}
