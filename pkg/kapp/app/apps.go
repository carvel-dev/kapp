package app

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	kappIsAppLabelKey   = "kapp.k14s.io/is-app"
	kappIsAppLabelValue = ""
	// TODO kappRevisionAnnKey   = "kapp.k14s.io/revision"
	// TODO kappLastDeployAnnKey = "kapp.k14s.io/last-deploy"
)

type Apps struct {
	nsName        string
	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

func NewApps(nsName string, coreClient kubernetes.Interface, dynamicClient dynamic.Interface) Apps {
	return Apps{nsName, coreClient, dynamicClient}
}

func (a Apps) Find(name string) (App, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("Expected app name to be non-empty")
	}

	const labelPrefix = "label:"

	if strings.HasPrefix(name, labelPrefix) {
		sel, err := labels.Parse(strings.TrimPrefix(name, labelPrefix))
		if err != nil {
			return nil, fmt.Errorf("Parsing app name (or label selector): %s", err)
		}
		return &LabeledApp{sel, a.coreClient, a.dynamicClient}, nil
	}

	return &RecordedApp{name, a.nsName, a.coreClient, a.dynamicClient, nil}, nil
}

func (a Apps) List(additionalLabels map[string]string) ([]App, error) {
	var result []App

	filterLabels := map[string]string{
		kappIsAppLabelKey: kappIsAppLabelValue,
	}

	for k, v := range additionalLabels {
		filterLabels[k] = v
	}

	listOpts := metav1.ListOptions{
		LabelSelector: labels.Set(filterLabels).String(),
	}

	apps, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).List(listOpts)
	if err != nil {
		return nil, err
	}

	for _, app := range apps.Items {
		meta := NewAppMetaFromData(app.Data)
		result = append(result, &RecordedApp{app.Name, a.nsName, a.coreClient, a.dynamicClient, &meta})
	}

	return result, nil
}
