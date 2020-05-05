package config

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

const defaultConfigYAML = `---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
# Copy over all metadata (with resourceVersion, etc.)
- path: [metadata]
  type: copy
  sources: [existing]
  resourceMatchers:
  - allMatcher: {}

# Be specific about labels to be applied
- path: [metadata, labels]
  type: remove
  resourceMatchers:
  - allMatcher: {}
- path: [metadata, labels]
  type: copy
  sources: [new]
  resourceMatchers:
  - allMatcher: {}

# Be specific about annotations to be applied
- path: [metadata, annotations]
  type: remove
  resourceMatchers:
  - allMatcher: {}
- path: [metadata, annotations]
  type: copy
  sources: [new]
  resourceMatchers:
  - allMatcher: {}

# Copy over all status, since cluster owns that
- path: [status]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - allMatcher: {}

# Prefer user provided, but allow cluster set
- path: [spec, clusterIP]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: Service

# Prefer user provided, but allow cluster set
- path: [spec, finalizers]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: Namespace

# Prefer user provided, but allow cluster set
- path: [secrets]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: ServiceAccount

# PVC
- path: [metadata, annotations, pv.kubernetes.io/bind-completed]
  type: copy
  sources: [new, existing]
  resourceMatchers: &pvcs
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: PersistentVolumeClaim

- path: [metadata, annotations, pv.kubernetes.io/bound-by-controller]
  type: copy
  sources: [new, existing]
  resourceMatchers: *pvcs

- path: [metadata, annotations, volume.beta.kubernetes.io/storage-provisioner]
  type: copy
  sources: [new, existing]
  resourceMatchers: *pvcs

- path: [spec, storageClassName]
  type: copy
  sources: [new, existing]
  resourceMatchers: *pvcs

- path: [spec, volumeMode]
  type: copy
  sources: [new, existing]
  resourceMatchers: *pvcs

- path: [spec, volumeName]
  type: copy
  sources: [new, existing]
  resourceMatchers: *pvcs

- path: [metadata, annotations, "deployment.kubernetes.io/revision"]
  type: copy
  sources: [new, existing]
  resourceMatchers: &appsV1DeploymentWithRevAnnKey
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: apps/v1beta2, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: Deployment}

- path: [webhooks, {allIndexes: true}, clientConfig, caBundle]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: admissionregistration.k8s.io/v1beta1, kind: MutatingWebhookConfiguration}
  - apiVersionKindMatcher: {apiVersion: admissionregistration.k8s.io/v1, kind: MutatingWebhookConfiguration}
  - apiVersionKindMatcher: {apiVersion: admissionregistration.k8s.io/v1beta1, kind: ValidatingWebhookConfiguration}
  - apiVersionKindMatcher: {apiVersion: admissionregistration.k8s.io/v1, kind: ValidatingWebhookConfiguration}

- path: [spec, caBundle]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: apiregistration.k8s.io/v1beta1, kind: APIService}
  - apiVersionKindMatcher: {apiVersion: apiregistration.k8s.io/v1, kind: APIService}

- path: [spec, nodeName]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}

diffAgainstLastAppliedFieldExclusionRules:
- path: [metadata, annotations, "deployment.kubernetes.io/revision"]
  resourceMatchers: *appsV1DeploymentWithRevAnnKey

diffMaskRules:
- path: [data]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Secret}
- path: [stringData]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Secret}

ownershipLabelRules:
- path: [metadata, labels]
  resourceMatchers:
  - allMatcher: {}

- path: [spec, template, metadata, labels]
  resourceMatchers: &withPodTemplate
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
  # Job
  - apiVersionKindMatcher: {apiVersion: batch/v1, kind: Job}

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
  resourceMatchers: *withPodTemplate

- path: [spec, selector, matchLabels]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: policy/v1beta1, kind: PodDisruptionBudget}

templateRules:
- resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: ConfigMap}
  affectedResources:
    objectReferences:
    - path: [spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, configMapKeyRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, configMapRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, initContainers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, configMapKeyRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, initContainers, {allIndexes: true}, envFrom, {allIndexes: true}, configMapRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, volumes, {allIndexes: true}, projected, sources, {allIndexes: true}, configMap]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, volumes, {allIndexes: true}, configMap]
      resourceMatchers: *withPodTemplate
    - path: [spec, volumes, {allIndexes: true}, configMap]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}

- resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Secret}
  affectedResources:
    objectReferences:
    - path: [spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, secretKeyRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, secretRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, initContainers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, secretKeyRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, initContainers, {allIndexes: true}, envFrom, {allIndexes: true}, secretRef]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, imagePullSecrets, {allIndexes: true}]
      resourceMatchers: *withPodTemplate
    - path: [spec, template, spec, volumes, {allIndexes: true}, secret]
      resourceMatchers: *withPodTemplate
      nameKey: secretName
    - path: [spec, template, spec, volumes, {allIndexes: true}, projected, sources, {allIndexes: true}, secret]
      resourceMatchers: *withPodTemplate
    - path: [spec, volumes, {allIndexes: true}, secret]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}
      nameKey: secretName
    - path: [spec, imagePullSecrets, {allIndexes: true}]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}
    - path: [imagePullSecrets, {allIndexes: true}]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: ServiceAccount}
    - path: [secrets, {allIndexes: true}]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: ServiceAccount}

changeGroupBindings:
- name: change-groups.kapp.k14s.io/crds
  resourceMatchers: &crdMatchers
  - apiGroupKindMatcher: {kind: CustomResourceDefinition, apiGroup: apiextensions.k8s.io}

- name: change-groups.kapp.k14s.io/namespaces
  resourceMatchers: &namespaceMatchers
  - apiGroupKindMatcher: {kind: Namespace, apiGroup: ""}

- name: change-groups.kapp.k14s.io/storage-class
  resourceMatchers: &storageClassMatchers
  - apiVersionKindMatcher: {kind: StorageClass, apiVersion: storage/v1}
  - apiVersionKindMatcher: {kind: StorageClass, apiVersion: storage/v1beta1}

- name: change-groups.kapp.k14s.io/storage
  resourceMatchers: &storageMatchers
  - apiVersionKindMatcher: {kind: PersistentVolume, apiVersion: v1}
  - apiVersionKindMatcher: {kind: PersistentVolumeClaim, apiVersion: v1}

- name: change-groups.kapp.k14s.io/rbac
  resourceMatchers: &rbacMatchers
  - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: ClusterRoleBinding, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: ClusterRoleBinding, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: ClusterRoleBinding, apiVersion: rbac.authorization.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: Role, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: Role, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: Role, apiVersion: rbac.authorization.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: RoleBinding, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: RoleBinding, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: RoleBinding, apiVersion: rbac.authorization.k8s.io/v1beta1}

- name: change-groups.kapp.k14s.io/pod-related
  resourceMatchers: &podRelatedMatchers
  - apiVersionKindMatcher: {kind: NetworkPolicy, apiVersion: extensions/v1beta1}
  - apiVersionKindMatcher: {kind: NetworkPolicy, apiVersion: networking.k8s.io/v1}
  - apiVersionKindMatcher: {kind: ResourceQuota, apiVersion: v1}
  - apiVersionKindMatcher: {kind: LimitRange, apiVersion: v1}
  - apiVersionKindMatcher: {kind: PodSecurityPolicy, apiVersion: extensions/v1beta1}
  - apiVersionKindMatcher: {kind: PodSecurityPolicy, apiVersion: policy/v1beta1}
  - apiVersionKindMatcher: {kind: PodDisruptionBudget, apiVersion: policy/v1beta1}
  - apiVersionKindMatcher: {kind: PriorityClass, apiVersion: scheduling.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: PriorityClass, apiVersion: scheduling.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: PriorityClass, apiVersion: scheduling.k8s.io/v1}
  - apiVersionKindMatcher: {kind: RuntimeClass, apiVersion: node.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: RuntimeClass, apiVersion: node.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: ServiceAccount, apiVersion: v1}
  - apiVersionKindMatcher: {kind: Secret, apiVersion: v1}
  - apiVersionKindMatcher: {kind: ConfigMap, apiVersion: v1}
  # [Note]: Do not add Service into this group as it may
  # delay other resources with load balancer provisioning
  # - apiVersionKindMatcher: {kind: Service, apiVersion: v1}

changeRuleBindings:
- rules:
  # [Note]: insert CRDs before all other custom resources
  - "upsert after upserting change-groups.kapp.k14s.io/crds"
  resourceMatchers:
  - andMatcher:
      matchers:
      - customResourceMatcher: {}
      - notMatcher:
          matcher:
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-change-group-and-rules]

- rules:
  - "delete before deleting change-groups.kapp.k14s.io/crds"
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher:
            anyMatcher:
              matchers: *crdMatchers
      - notMatcher:
          matcher:
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-change-group-and-rules]

- rules:
  # [Note]: insert namespaces before all other namespaced resources
  - "upsert after upserting change-groups.kapp.k14s.io/namespaces"
  resourceMatchers:
  - andMatcher:
      matchers:
      - hasNamespaceMatcher: {}
      - notMatcher:
          matcher:
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-change-group-and-rules]

- rules:
  - "upsert after upserting change-groups.kapp.k14s.io/storage-class"
  ignoreIfCyclical: true
  resourceMatchers:
  - apiVersionKindMatcher: {kind: PersistentVolume, apiVersion: v1}
  - apiVersionKindMatcher: {kind: PersistentVolumeClaim, apiVersion: v1}

- rules:
  # [Note]: prefer to apply pod related changes first to
  # work better with applications that do not reload changes
  - "upsert after upserting change-groups.kapp.k14s.io/pod-related"
  # [Note]: prefer to apply rbac changes first to potentially
  # avoid restarts of Pods that rely on correct permissions
  - "upsert after upserting change-groups.kapp.k14s.io/rbac"
  - "upsert after upserting change-groups.kapp.k14s.io/storage-class"
  - "upsert after upserting change-groups.kapp.k14s.io/storage"
  ignoreIfCyclical: true
  resourceMatchers:
  # [Note]: Apply all resources after pod-related change group as it's
  # common for other resources to rely on ConfigMaps, Secrets, etc.
  - andMatcher:
      matchers:
      - notMatcher:
          matcher:
            anyMatcher:
              matchers:
              - anyMatcher: {matchers: *storageClassMatchers}
              - anyMatcher: {matchers: *storageMatchers}
              - anyMatcher: {matchers: *rbacMatchers}
              - anyMatcher: {matchers: *podRelatedMatchers}
      - hasNamespaceMatcher: {}
`

var defaultConfigRes = ctlres.MustNewResourceFromBytes([]byte(defaultConfigYAML))

func NewDefaultConfigString() string { return defaultConfigYAML }

func NewConfFromResourcesWithDefaults(resources []ctlres.Resource) ([]ctlres.Resource, Conf, error) {
	return NewConfFromResources(append([]ctlres.Resource{defaultConfigRes}, resources...))
}
