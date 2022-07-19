// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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
  sources: [existing]
  resourceMatchers:
  - allMatcher: {}

# Prefer user provided, but allow cluster set
- paths:
  - [spec, clusterIP]
  - [spec, healthCheckNodePort]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}

# Prefer user provided, but allow cluster set
- path: [spec, finalizers]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Namespace}

# Openshift adds some annotations and labels to namespaces
- paths:
  - [metadata, annotations, openshift.io/sa.scc.mcs]
  - [metadata, annotations, openshift.io/sa.scc.supplemental-groups]
  - [metadata, annotations, openshift.io/sa.scc.uid-range]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Namespace}

# PVC
- paths:
  - [metadata, annotations, pv.kubernetes.io/bind-completed]
  - [metadata, annotations, pv.kubernetes.io/bound-by-controller]
  - [metadata, annotations, volume.beta.kubernetes.io/storage-provisioner]
  - [spec, storageClassName]
  - [spec, volumeMode]
  - [spec, volumeName]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: PersistentVolumeClaim}

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

- path: [spec, conversion, webhookClientConfig, caBundle]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: apiextensions.k8s.io/v1beta1, kind: CustomResourceDefinition}

- path: [spec, conversion, webhook, clientConfig, caBundle]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: apiextensions.k8s.io/v1, kind: CustomResourceDefinition}

- path: [spec, nodeName]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Pod}

# ServiceAccount controller appends secret named '${metadata.name}-token-${rand}' after the save
# Openshift adds a secret and an imagePullSecret named '${metadata.name}-dockercfg-${rand}' after the save
- ytt:
    overlayContractV1:
      overlay.yml: |
        #@ load("@ytt:data", "data")
        #@ load("@ytt:overlay", "overlay")

        #@ res_name = data.values.existing.metadata.name

        #! service account may be created with empty secrets
        #@ secrets = []
        #@ if hasattr(data.values.existing, "secrets"):
        #@   secrets = data.values.existing.secrets
        #@ end

        #@ imagePullSecrets = []
        #@ if hasattr(data.values.existing, "imagePullSecrets"):
        #@   imagePullSecrets = data.values.existing.imagePullSecrets
        #@ end

        #@ token_secret_name = None
        #@ token_secret_name_docker = None
        #@ for k in secrets:
        #@   if k.name.startswith(res_name+"-token-"):
        #@     token_secret_name = k.name
        #@   end
        #@   if k.name.startswith(res_name+"-dockercfg-"):
        #@     token_secret_name_docker = k.name
        #@   end
        #@ end

        #@ image_pull_secret_name = None
        #@ for k in imagePullSecrets:
        #@   if k.name.startswith(res_name+"-dockercfg-"):
        #@     image_pull_secret_name = k.name
        #@   end
        #@ end

        #! in case token secret name is not included, do not modify anything

        #@ if/end image_pull_secret_name:
        #@overlay/match by=overlay.all
        ---
        #@overlay/match missing_ok=True
        imagePullSecrets:
        #@overlay/match by=overlay.subset({"name": image_pull_secret_name}),when=0
        - name: #@ image_pull_secret_name

        #@ if/end token_secret_name:
        #@overlay/match by=overlay.all
        ---
        #@overlay/match missing_ok=True
        secrets:
        #@overlay/match by=overlay.subset({"name": token_secret_name}),when=0
        - name: #@ token_secret_name

        #@ if/end token_secret_name_docker:
        #@overlay/match by=overlay.all
        ---
        #@overlay/match missing_ok=True
        secrets:
        #@overlay/match by=overlay.subset({"name": token_secret_name_docker}),when=0
        - name: #@ token_secret_name_docker
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: ServiceAccount}

# Secretgen populates secret data for annotated secrets
- paths:
  - [data, .dockerconfigjson]
  - [metadata, annotations, secretgen.carvel.dev/status]
  type: copy
  sources: [existing, new]
  resourceMatchers:
  - andMatcher:
      matchers:
      - apiVersionKindMatcher: {apiVersion: v1, kind: Secret}
      - hasAnnotationMatcher:
          keys: [secretgen.carvel.dev/image-pull-secret]
      - notMatcher:
          matcher:
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-secretgen-rebase-rules]

# aggregated ClusterRole rules are filled in by the control plane at runtime
# refs https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles
- paths:
  - [rules]
  type: copy
  sources: [existing]
  resourceMatchers:
  - andMatcher:
      matchers:
      - anyMatcher:
          matchers:
          - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1}
          - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1alpha1}
          - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1beta1}
      - notMatcher:
          matcher:
            emptyFieldMatcher:
              path: [aggregationRule]

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
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher: &disableDefaultOwnershipLabelRulesAnnMatcher
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-ownership-label-rules]
      - anyMatcher:
          matchers: &withPodTemplate
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
  - andMatcher:
      matchers:
      - notMatcher:
          matcher: *disableDefaultOwnershipLabelRulesAnnMatcher
      - anyMatcher:
          matchers:
          # StatefulSet
          - apiVersionKindMatcher: {apiVersion: apps/v1, kind: StatefulSet}
          - apiVersionKindMatcher: {apiVersion: apps/v1beta1, kind: StatefulSet}
          - apiVersionKindMatcher: {apiVersion: extensions/v1beta1, kind: StatefulSet}

- path: [spec, template, metadata, labels]
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher: *disableDefaultOwnershipLabelRulesAnnMatcher
      - anyMatcher:
          matchers:
          - apiVersionKindMatcher: {apiVersion: batch/v1, kind: Job}
          - apiVersionKindMatcher: {apiVersion: batch/v1beta1, kind: Job}
          - apiVersionKindMatcher: {apiVersion: batch/v2alpha1, kind: Job}

- path: [spec, jobTemplate, spec, template, metadata, labels]
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher: *disableDefaultOwnershipLabelRulesAnnMatcher
      - anyMatcher:
          matchers: &cronJob
          - apiVersionKindMatcher: {apiVersion: batch/v1, kind: CronJob}
          - apiVersionKindMatcher: {apiVersion: batch/v1beta1, kind: CronJob}
          - apiVersionKindMatcher: {apiVersion: batch/v2alpha1, kind: CronJob}

labelScopingRules:
- path: [spec, selector]
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          # Keep older annotation for backwards compatibility
          matcher: &disableLabelScopingAnnMatcher
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-label-scoping]
      - notMatcher:
          matcher: &disableDefaultLabelScopingRulesAnnMatcher
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-label-scoping-rules]
      - apiVersionKindMatcher: {apiVersion: v1, kind: Service}

- path: [spec, selector, matchLabels]
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher: *disableLabelScopingAnnMatcher
      - notMatcher:
          matcher: *disableDefaultLabelScopingRulesAnnMatcher
      - anyMatcher:
          matchers: *withPodTemplate

- path: [spec, selector, matchLabels]
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher: *disableLabelScopingAnnMatcher
      - notMatcher:
          matcher: *disableDefaultLabelScopingRulesAnnMatcher
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

    - path: [spec, jobTemplate, spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, configMapKeyRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, configMapRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, initContainers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, configMapKeyRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, initContainers, {allIndexes: true}, envFrom, {allIndexes: true}, configMapRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, volumes, {allIndexes: true}, projected, sources, {allIndexes: true}, configMap]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, volumes, {allIndexes: true}, configMap]
      resourceMatchers: *cronJob

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

    - path: [spec, jobTemplate, spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, secretKeyRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, secretRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, initContainers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, secretKeyRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, initContainers, {allIndexes: true}, envFrom, {allIndexes: true}, secretRef]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, imagePullSecrets, {allIndexes: true}]
      resourceMatchers: *cronJob
    - path: [spec, jobTemplate, spec, template, spec, volumes, {allIndexes: true}, secret]
      resourceMatchers: *cronJob
      nameKey: secretName
    - path: [spec, jobTemplate, spec, template, spec, volumes, {allIndexes: true}, projected, sources, {allIndexes: true}, secret]
      resourceMatchers: *cronJob

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

- name: change-groups.kapp.k14s.io/crds-{crd-group}-{crd-kind}
  resourceMatchers: *crdMatchers

- name: change-groups.kapp.k14s.io/namespaces
  resourceMatchers: &namespaceMatchers
  - apiGroupKindMatcher: {kind: Namespace, apiGroup: ""}

- name: change-groups.kapp.k14s.io/namespaces-{name}
  resourceMatchers: *namespaceMatchers

- name: change-groups.kapp.k14s.io/storage-class
  resourceMatchers: &storageClassMatchers
  - apiVersionKindMatcher: {kind: StorageClass, apiVersion: storage/v1}
  - apiVersionKindMatcher: {kind: StorageClass, apiVersion: storage/v1beta1}

- name: change-groups.kapp.k14s.io/storage
  resourceMatchers: &storageMatchers
  - apiVersionKindMatcher: {kind: PersistentVolume, apiVersion: v1}
  - apiVersionKindMatcher: {kind: PersistentVolumeClaim, apiVersion: v1}

- name: change-groups.kapp.k14s.io/rbac-roles
  resourceMatchers: &rbacRoleMatchers
  - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: ClusterRole, apiVersion: rbac.authorization.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: Role, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: Role, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: Role, apiVersion: rbac.authorization.k8s.io/v1beta1}
  
- name: change-groups.kapp.k14s.io/rbac-role-bindings
  resourceMatchers: &rbacRoleBindingMatchers
  - apiVersionKindMatcher: {kind: ClusterRoleBinding, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: ClusterRoleBinding, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: ClusterRoleBinding, apiVersion: rbac.authorization.k8s.io/v1beta1}
  - apiVersionKindMatcher: {kind: RoleBinding, apiVersion: rbac.authorization.k8s.io/v1}
  - apiVersionKindMatcher: {kind: RoleBinding, apiVersion: rbac.authorization.k8s.io/v1alpha1}
  - apiVersionKindMatcher: {kind: RoleBinding, apiVersion: rbac.authorization.k8s.io/v1beta1}

- name: change-groups.kapp.k14s.io/rbac
  resourceMatchers: &rbacMatchers
  - anyMatcher:
      matchers:
        - anyMatcher: {matchers: *rbacRoleMatchers}
        - anyMatcher: {matchers: *rbacRoleBindingMatchers}

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

- name: change-groups.kapp.k14s.io/serviceaccount
  resourceMatchers: &serviceAccountMatchers
  - apiVersionKindMatcher: {kind: ServiceAccount, apiVersion: v1}

- name: change-groups.kapp.k14s.io/appcr
  resourceMatchers:
  - apiVersionKindMatcher: {kind: App, apiVersion: kappctrl.k14s.io/v1alpha1}

- name: change-groups.kapp.k14s.io/packageinstall
  resourceMatchers:
  - apiVersionKindMatcher: {kind: PackageInstall, apiVersion: packaging.carvel.dev/v1alpha1}

changeRuleBindings:
# Insert CRDs before all CRs
- rules:
  - "upsert after upserting change-groups.kapp.k14s.io/crds-{api-group}-{kind}"
  resourceMatchers:
  - andMatcher:
      matchers:
      - customResourceMatcher: {}
      - notMatcher:
          matcher: &disableDefaultChangeGroupAnnMatcher
            hasAnnotationMatcher:
              keys: [kapp.k14s.io/disable-default-change-group-and-rules]

# Delete CRs before CRDs to retain detailed observability
# instead of having CRD deletion trigger all CR deletion
- rules:
  - "delete before deleting change-groups.kapp.k14s.io/crds"
  ignoreIfCyclical: true
  resourceMatchers:
  - andMatcher:
      matchers:
      - customResourceMatcher: {}
      - notMatcher:
          matcher: *disableDefaultChangeGroupAnnMatcher

# Delete non-CRs after deleting CRDs so that if CRDs use conversion
# webhooks it's more likely that backing webhook resources are still
# available during deletion of CRs
- rules:
  - "delete after deleting change-groups.kapp.k14s.io/crds"
  ignoreIfCyclical: true
  resourceMatchers:
  - andMatcher:
      matchers:
      - notMatcher:
          matcher:
            customResourceMatcher: {}
      - notMatcher:
          matcher:
            anyMatcher:
              matchers: *crdMatchers
      - notMatcher:
          matcher: *disableDefaultChangeGroupAnnMatcher

# Insert namespaces before all namespaced resources
- rules:
  - "upsert after upserting change-groups.kapp.k14s.io/namespaces-{namespace}"
  resourceMatchers:
  - andMatcher:
      matchers:
      - hasNamespaceMatcher: {}
      - notMatcher:
          matcher: *disableDefaultChangeGroupAnnMatcher

# Insert roles/ClusterRoles before inserting any roleBinding/ClusterRoleBinding
# Sometimes Binding Creation fail as corresponding Role is not created.
# https://github.com/vmware-tanzu/carvel-kapp/issues/145
- rules:
  - "upsert after upserting change-groups.kapp.k14s.io/rbac-roles"
  ignoreIfCyclical: true
  resourceMatchers:
  - andMatcher:
      matchers:
      - anyMatcher: {matchers: *rbacRoleBindingMatchers}
      - notMatcher:
          matcher: *disableDefaultChangeGroupAnnMatcher

- rules:
  - "upsert before upserting change-groups.kapp.k14s.io/packageinstall"
  - "upsert before upserting change-groups.kapp.k14s.io/appcr"
  - "delete after deleting change-groups.kapp.k14s.io/packageinstall"
  - "delete after deleting change-groups.kapp.k14s.io/appcr"
  ignoreIfCyclical: true
  resourceMatchers:
  - anyMatcher: {matchers: *serviceAccountMatchers}
  - anyMatcher: {matchers: *rbacMatchers}

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
