## Config

kapp supports custom `Config` resource to specify its own configuration. Config resource is never applied to the cluster, though it follows general Kubernetes resource format. Multiple config resources are allowed.

kapp comes with __built-in configuration__ (see it via `kapp deploy-config`) that includes rules for common resources.

### Format

```yaml
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

minimumRequiredVersion: 0.23.0

rebaseRules:
- path: [spec, clusterIP]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}

ownershipLabelRules:
- path: [metadata, labels]
  resourceMatchers:
  - allMatcher: {}

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
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}

diffMaskRules:
- path: [data]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Secret}
```

#### minimumRequiredVersion

`minimumRequiredVersion` forces kapp to exit with a validation error if kapp's version is below minimum required version. Available in v0.23.0+.

#### rebaseRules

`rebaseRules` specify origin of field values. Kubernetes cluster generates (or defaults) some field values, hence these values will need to be merged in future to avoid flagging them during diffing. Common example is `v1/Service`'s `spec.clusterIP` field is automatically populated if it's not set. See [HPA and Deployment rebase](hpa-deployment-rebase.md) or [PersistentVolumeClaim rebase](rebase-pvc.md) examples.

- `rebaseRules` (array) list of rebase rules
  - `path` (array) specifies location within a resource to rebase. Mutually exclusive with `paths`. Example: `[spec, clusterIP]`
  - `paths` (array of paths) specifies multiple locations within a resource to rebase. This is a convenience for specifying multiple rebase rules with only different paths. Mutually exclusive with `path`. Available in v0.27.0+.
  - `sources` (array of strings) specifies source preference from where to copy value from. Allowed values: `new` or `existing`. Example: `[new, existing]`
  - `resourceMatchers` (array) specifies rules to find matching resources. See various resource matchers below.

```yaml
rebaseRules:
- path: [spec, clusterIP]
  type: copy
  sources: [new, existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}
```

```yaml
rebaseRules:
- paths:
  - [spec, clusterIP]
  - [spec, healthCheckNodePort]
  type: copy
  sources: [existing]
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: v1, kind: Service}
```

#### ownershipLabelRules

`ownershipLabelRules` specify locations for inserting kapp generated labels. These labels allow kapp to track which resources belong to which application. For resources that describe creation of other resources (e.g. `Deployment` or `StatefulSet`), configuration may need to specify where to insert labels for child resources that will be created.

#### labelScopingRules

`labelScopingRules` specify locations for inserting kapp generated labels that scope resources to resources within current application. `kapp.k14s.io/disable-label-scoping: ""` (value must be empty) annotation can be used to exclude an individual resource from label scoping.

#### waitRules

Available in v0.29.0+.

`waitRules` specify how to wait for resources that kapp does not wait for by default. Each rule provides a way to specify which `status.conditions` indicate success or failure. Once a condition matches the status' value kapp will exit waiting and report the result. (If this functionality is not enough to wait for resources in your use case, please reach out on Slack to discuss further.)

```yaml
waitRules:
- supportsObservedGeneration: true
  conditionMatchers:
  - type: Failed
    status: "True"
    failure: true
  - type: Deployed
    status: "True"
    success: true
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: corp.com/v1, kind: DatabaseInstance}
```

```yaml
waitRules:
- supportsObservedGeneration: true
  conditionMatchers:
  - type: Ready
    status: "False"
    failure: true
  - type: Ready
    status: "True"
    success: true
  resourceMatchers:
  - apiVersionKindMatcher: {apiVersion: corp.com/v1, kind: Application}
```

#### templateRules

`templateRules` how template resources affect other resources. In above example, template config maps are said to affect deployments.

#### additionalLabels

`additionalLabels` specify additional labels to apply to all resources for custom uses by the user (added based on `ownershipLabelRules`).

#### diffAgainstLastAppliedFieldExclusionRules

`diffAgainstLastAppliedFieldExclusionRules` specify which fields should be removed before diff-ing against last applied resource. These rules are useful for fields are "owned" by the cluster/controllers, and are only later updated. For example `Deployment` resource has an annotation that gets set after a little bit of time after resource is created/updated (not during resource admission). It's typically not necessary to use this configuration.

#### diffMaskRules

`diffMaskRules` specify which field values should be masked in diff. By default `v1/Secret`'s `data` fields are masked. Currently only applied to `deploy` command.

#### changeGroupBindings

Available in v0.25.0+.

`changeGroupBindings` bind specified change group to resources matched by resource matchers. This is an alternative to using `kapp.k14s.io/change-group` annotation to add change group to resources. See `kapp deploy-config` for default bindings.

#### changeRuleBindings

Available in v0.25.0+.

`changeRuleBindings` bind specified change rules to resources matched by resource matchers. This is an alternative to using `kapp.k14s.io/change-rule` annotation to add change rules to resources. See `kapp deploy-config` for default bindings.

---
### Resource matchers

Resource matchers (as used by `rebaseRules` and `ownershipLabelRules`):

#### all

Matches all resources

```yaml
allMatcher: {}
```

#### any

Matches resources that match one of matchers

```yaml
anyMatcher:
  matchers:
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
  - apiVersionKindMatcher: {apiVersion: extensions/v1alpha1, kind: Deployment}
```

#### not

Matches any resource that does not match given matcher

```yaml
notMatcher:
  matcher:
    apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
```

#### and

Matches any resource that matches all given matchers

```yaml
andMatcher:
  matchers:
  - apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
  - hasNamespaceMatcher: {}
```

#### apiGroupKindMatcher

```yaml
apiVersionKindMatcher: {apiGroup: apps, kind: Deployment}
```

#### apiVersionKindMatcher

```yaml
apiVersionKindMatcher: {apiVersion: apps/v1, kind: Deployment}
```

#### kindNamespaceNameMatcher

```yaml
kindNamespaceNameMatcher: {kind: Deployment, namespace: mysql, name: mysql}
```

#### hasAnnotationMatcher

Matches resources that have particular annotation

```yaml
hasAnnotationMatcher:
  keys:
  - kapp.k14s.io/change-group
```

#### hasNamespaceMatcher

Matches any resource that has a non-empty namespace

```yaml
hasNamespaceMatcher: {}
```

Matches any resource with namespace that equals to one of the specified names

```yaml
hasNamespaceMatcher:
  names: [app1, app2]
```

#### customResourceMatcher

Matches any resource that is not part of builtin k8s API groups (e.g. apps, batch, etc.). It's likely that over time some builtin k8s resources would not be matched.

```yaml
customResourceMatcher: {}
```

---
### Paths

Path specifies location within a resource (as used `rebaseRules` and `ownershipLabelRules`):

```
[spec, clusterIP]
```

```
[spec, volumeClaimTemplates, {allIndexes: true}, metadata, labels]
```

```
[spec, volumeClaimTemplates, {index: 0}, metadata, labels]
```
