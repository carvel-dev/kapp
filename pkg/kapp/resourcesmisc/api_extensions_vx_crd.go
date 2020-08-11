/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package resourcesmisc

import (
	"github.com/ghodss/yaml"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ApiExtensionsVxCRD struct {
	resource ctlres.Resource
}

func NewApiExtensionsVxCRD(resource ctlres.Resource) *ApiExtensionsVxCRD {
	matcher := ctlres.APIGroupKindMatcher{
		APIGroup: "apiextensions.k8s.io",
		Kind:     "CustomResourceDefinition",
	}
	if matcher.Matches(resource) {
		return &ApiExtensionsVxCRD{resource}
	}
	return nil
}

func (s ApiExtensionsVxCRD) IsDoneApplying() DoneApplyState {
	// CRD conditions: https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions
	allTrue, msg := Conditions{s.resource}.IsSelectedTrue([]string{"Established", "NamesAccepted"})
	return DoneApplyState{Done: allTrue, Successful: allTrue, Message: msg}
}

func (s ApiExtensionsVxCRD) contents() (crdObj, error) {
	bs, err := s.resource.AsYAMLBytes()
	if err != nil {
		return crdObj{}, err
	}

	var contents crdObj

	err = yaml.Unmarshal(bs, &contents)
	if err != nil {
		return crdObj{}, err
	}

	return contents, nil
}

// TODO use struct provided by the client
type crdObj struct {
	Spec crdSpec `yaml:"spec"`
}

type crdSpec struct {
	Group    string           `yaml:"group"`
	Scope    string           `yaml:"scope"`
	Version  string           `yaml:"version"`
	Versions []crdSpecVersion `yaml:"versions"`
	Names    crdSpecNames     `yaml:"names"`
}

type crdSpecVersion struct {
	Name string `yaml:"name"`
}

type crdSpecNames struct {
	Kind string `yaml:"kind"`
}

func (o crdObj) Versions() []string {
	result := []string{}

	if len(o.Spec.Version) > 0 {
		result = append(result, o.Spec.Version)
	}
	if len(o.Spec.Versions) > 0 {
		for _, ver := range o.Spec.Versions {
			result = append(result, ver.Name)
		}
	}

	return result
}

/*

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: builds.build.knative.dev
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].status
    name: Succeeded
    type: string
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].reason
    name: Reason
    type: string
  - JSONPath: .status.startTime
    name: StartTime
    type: date
  - JSONPath: .status.completionTime
    name: CompletionTime
    type: date
  group: build.knative.dev
  names:
    categories:
    - all
    - knative
    kind: Build
    plural: builds
  scope: Namespaced
  version: v1alpha1
status:
  conditions:
  - lastTransitionTime: 2018-12-06T02:02:55Z
    message: no conflicts found
    reason: NoConflicts
    status: "True"
    type: NamesAccepted
  - lastTransitionTime: 2018-12-06T02:02:55Z
    message: the initial names have been accepted
    reason: InitialNamesAccepted
    status: "True"
    type: Established

*/
