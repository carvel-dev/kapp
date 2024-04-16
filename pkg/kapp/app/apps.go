// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"carvel.dev/kapp/pkg/kapp/logger"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	KappIsAppLabelKey   = "kapp.k14s.io/is-app"
	kappIsAppLabelValue = ""
)

type Apps struct {
	nsName              string
	coreClient          kubernetes.Interface
	identifiedResources ctlres.IdentifiedResources
	logger              logger.Logger
}

func NewApps(nsName string, coreClient kubernetes.Interface,
	identifiedResources ctlres.IdentifiedResources, logger logger.Logger) Apps {

	return Apps{nsName, coreClient, identifiedResources, logger}
}

func (a Apps) Find(name string) (App, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("Expected app name to be non-empty")
	}

	const labelPrefix = "label:"

	if strings.HasPrefix(name, labelPrefix) {
		sel, err := labels.Parse(strings.TrimPrefix(name, labelPrefix))
		if err != nil {
			return nil, fmt.Errorf("Parsing app name (or label selector): %w", err)
		}
		return &LabeledApp{sel, a.identifiedResources}, nil
	}

	if len(a.nsName) == 0 {
		return nil, fmt.Errorf("Expected non-empty namespace")
	}

	return NewRecordedApp(name, a.nsName, time.Time{}, a.coreClient, a.identifiedResources, a.appInDiffNsHintMsg, a.logger), nil
}

func (a Apps) List(additionalLabels map[string]string) ([]App, error) {
	return a.list(additionalLabels, a.nsName)
}

func (a Apps) list(additionalLabels map[string]string, nsName string) ([]App, error) {
	var result []App

	filterLabels := map[string]string{
		KappIsAppLabelKey: kappIsAppLabelValue,
	}

	for k, v := range additionalLabels {
		filterLabels[k] = v
	}

	listOpts := metav1.ListOptions{
		LabelSelector: labels.Set(filterLabels).String(),
	}

	apps, err := a.coreClient.CoreV1().ConfigMaps(nsName).List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}

	for _, app := range apps.Items {
		recordedApp := NewRecordedApp(app.Name, app.Namespace, app.ObjectMeta.CreationTimestamp.Time, a.coreClient,
			a.identifiedResources, a.appInDiffNsHintMsg, a.logger)

		recordedApp.setMeta(app)

		result = append(result, recordedApp)
	}

	return result, nil
}

func (a Apps) appInDiffNsHintMsg(name string) string {
	items, err := a.list(nil, "")
	if err != nil {
		return ""
	}

	var foundNss []string

	for _, item := range items {
		if item.Name() == name {
			foundNss = append(foundNss, item.Namespace())
		}
	}

	if len(foundNss) > 0 {
		if len(foundNss) > 3 {
			foundNss = append(foundNss[:3], "...")
		}
		return fmt.Sprintf(" (hint: found app '%s' in other namespaces: %s)",
			name, strings.Join(foundNss, ", "))
	}
	return ""
}
