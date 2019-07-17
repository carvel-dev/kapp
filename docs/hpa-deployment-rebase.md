## HPA and Deployment rebase

Here is an example on how to use custom `rebaseRules` to "prefer" server chosen value for `spec.replicas` field for a particular Deployment.

```yaml
apiVersion: kapp.k14s.io/v1alpha1
kind: Config
rebaseRules:
- path: [spec, replicas]
  type: copy
  sources: [existing, new]
  resourceMatchers:
  - kindNamespaceNameMatcher:
      kind: Deployment
      namespace: my-ns
      name: my-app
---
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: my-app
  namespace: my-ns
...
---
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: my-app
  namespace: my-ns
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 50
```
