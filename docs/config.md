## Config

kapp supports custom `Config` resource to specify its own configuration. Config resource is never applied to the cluster, though it follows general Kubernetes resource format. Multiple config resources are allowed.

kapp comes with __built-in configuration__ (see it via `kapp deploy-config`) that includes rules for common resources.

### Format

```yaml
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
- path: [spec, clusterIP]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: Service

ownershipLabelRules:
- path: [metadata, labels]
  resourceMatchers:
  - allResourceMatcher: {}

labelScopingRules:
- path: [spec, selector]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}

templateRules:
- resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: ConfigMap}
  affectedResources:
    objectReferences:
    - path: [spec, template, spec, containers, {allIndexes: true}, env, {allIndexes: true}, valueFrom, configMapKeyRef]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
    - path: [spec, template, spec, containers, {allIndexes: true}, envFrom, {allIndexes: true}, configMapRef]
      resourceMatchers:
      - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}

additionalLabels:
  department: marketing
  cost-center: mar201

diffAgainstLastAppliedFieldExclusionRules:
- path: [metadata, annotations, "deployment.kubernetes.io/revision"]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: apps/v1
      kind: Deployment

diffMaskRules:
- path: [data]
  resourceMatchers:
  - apiVersionKindMatcher:
      apiVersion: v1
      kind: Secret
```

`rebaseRules` specify origin of field values. Kubernetes cluster generates (or defaults) some field values, hence these values will need to be merged in future to avoid flagging them during diffing. Common example is `v1/Service`'s `spec.clusterIP` field is automatically populated if it's not set. See [HPA and Deployment rebase](hpa-deployment-rebase.md) or [PersistentVolumeClaim rebase](rebase-pvc.md) examples.

`ownershipLabelRules` specify locations for inserting kapp generated labels. These labels allow kapp to track which resources belong to which application. For resources that describe creation of other resources (e.g. `Deployment` or `StatefulSet`), configuration may need to specify where to insert labels for child resources that will be created.

`labelScopingRules` specify locations for inserting kapp generated labels that scope resources to resources within current application. `kapp.k14s.io/disable-label-scoping: ""` (value must be empty) annotation can be used to exclude an individual resource from label scoping.

`templateRules` how template resources affect other resources. In above example, template config maps are said to affect deployments.

`additionalLabels` specify additional labels to apply to all resources for custom uses by the user (added based on `ownershipLabelRules`).

`diffAgainstLastAppliedFieldExclusionRules` specify which fields should be removed before diff-ing against last applied resource. These rules are useful for fields are "owned" by the cluster/controllers, and are only later updated. For example `Deployment` resource has an annotation that gets set after a little bit of time after resource is created/updated (not during resource admission). It's typically not necessary to use this configuration.

`diffMaskRules` specify which field values should be masked in diff. By default `v1/Secret`'s `data` fields are masked. Currently only applied to `deploy` command.

### Resource matchers

Resource matchers (as used by `rebaseRules` and `ownershipLabelRules`):

```yaml
allResourceMatcher: {}
```

```yaml
apiVersionKindMatcher:
  APIVersion: apps/v1
  kind: Deployment
```

```yaml
kindNamespaceNameMatcher:
  kind: Deployment
  namespace: mysql
  name: mysql
```

### Paths

Path specifies location within a resource (as used `rebaseRules` and `ownershipLabelRules`):

```yaml
[spec, clusterIP]
```

```yaml
[spec, volumeClaimTemplates, {allIndexes: true}, metadata, labels]
```

```yaml
[spec, volumeClaimTemplates, {index: 0}, metadata, labels]
```
