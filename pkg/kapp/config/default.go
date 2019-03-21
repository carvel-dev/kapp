package config

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const defaultConfigYAML = `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
# Copy over all metadata (with reosurceVersion, etc.)
- path: [metadata]
  merge: copy
  sources: [existing]
  resourceMatchers:
  - allResourceMatcher: {}

# Be specific about labels to be applied
- path: [metadata, labels]
  merge: copy
  sources: [new]
  resourceMatchers:
  - allResourceMatcher: {}

# Be specific about annotations to be applied
- path: [metadata, annotations]
  merge: copy
  sources: [new]
  resourceMatchers:
  - allResourceMatcher: {}

# Copy over all status, since cluster owns that
- path: [status]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - allResourceMatcher: {}

# Prefer user provided, but allow cluster set
- path: [spec, clusterIP]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: Service

# Prefer user provided, but allow cluster set
- path: [spec, finalizers]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: Namespace

# Prefer user provided, but allow cluster set
- path: [secrets]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: ServiceAccount

- path: [spec, storageClassName]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: PersistentVolumeClaim

- path: [spec, volumeName]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: PersistentVolumeClaim

- path: [metadata, annotations, "deployment.kubernetes.io/revision"]
  merge: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: apps/v1
      kind: Deployment
  - apiVersionKindMatcher:
      apiVersion: extensions/v1beta1
      kind: Deployment

ownershipLabelRules:
- path: [metadata, labels]
  resourceMatchers:
  - allResourceMatcher: {}

- path: [spec, template, metadata, labels]
  resourceMatchers: &builtinAppsControllers
  # Deployment
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta2, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: Deployment}
  # ReplicaSet
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: ReplicaSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta2, kind: ReplicaSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: ReplicaSet}
  - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: ReplicaSet}
  # StatefulSet
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: StatefulSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta2, kind: StatefulSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: StatefulSet}
  - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: StatefulSet}
  # DaemonSet
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: DaemonSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta2, kind: DaemonSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: DaemonSet}
  - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: DaemonSet}

# TODO It seems that these labels are being ignored
# https://github.com/kubernetes/kubernetes/issues/74916
- path: [spec, volumeClaimTemplates, {allIndexes: true}, metadata, labels]
  resourceMatchers:
  # StatefulSet
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: StatefulSet}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: StatefulSet}
  - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: StatefulSet}

- path: [spec, template, metadata, labels]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: batch/v1, kind: Job}
  - apiVersionKindMatcher: {apiVersion: batch/v1beta1, kind: Job}
  - apiVersionKindMatcher: {apiVersion: batch/v2alpha1, kind: Job}

- path: [spec, jobTemplate, spec, template, metadata, labels]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: batch/v1beta1, kind: CronJob}
  - apiVersionKindMatcher: {apiVersion: batch/v2alpha1, kind: CronJob}

labelScopingRules:
- path: [spec, selector]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}

- path: [spec, selector, matchLabels]
  resourceMatchers: *builtinAppsControllers

- path: [spec, selector, matchLabels]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: policy/v1beta1, kind: PodDisruptionBudget}

templateRules:
- resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: ConfigMap}
  affectedResources:
    objectReferences:
    - path: [spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, configMapKeyRef]
      resourceMatchers: *builtinAppsControllers
    - path: [spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, configMapRef]
      resourceMatchers: *builtinAppsControllers
    - path: [spec, template, spec, volumes, {allIndexes: true}, configMap]
      resourceMatchers: *builtinAppsControllers
    - path: [spec, volumes, {allIndexes: true}, configMap]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}

- resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Secret}
  affectedResources:
    objectReferences:
    - path: [spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, secretKeyRef]
      resourceMatchers: *builtinAppsControllers
    - path: [spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, secretRef]
      resourceMatchers: *builtinAppsControllers
    # TODO uses secretName instead of name
    #- path: [spec, template, spec, volumes, {allIndexes: true}, secret]
    #  resourceMatchers: *builtinAppsControllers
    #- path: [spec, volumes, {allIndexes: true}, secret]
    #  resourceMatchers:
    #  - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}
`

var defaultConfigRes = ctlres.MustNewResourceFromBytes([]byte(defaultConfigYAML))

func NewDefaultConfigString() string { return defaultConfigYAML }

func NewConfFromResourcesWithDefaults(resources []ctlres.Resource) ([]ctlres.Resource, Conf, error) {
	return NewConfFromResources(append([]ctlres.Resource{defaultConfigRes}, resources...))
}
