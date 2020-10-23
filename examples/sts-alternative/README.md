# STS alternative

StatefulSets have various surprising behaviours. This set of templates shows how to implement common STS feature set in a very lightweight manner with a little bit of ytt temlates and kapp features.

```
kapp deploy -a redis -f <(ytt -f examples/sts-alternative)
```

Highlights:

- creates 3 Pods (redis-0, ...) are created with 3 PVCs (redis-0, ...)
- ordered rollout of changes is controlled via kapp change rules
  - example: updated Pod 0, then Pod 1, then Pod 2
- Pods are recreated when ConfigMap changes
  - ConfigMap is marked as versioned resource (via `kapp.k14s.io/versioned` annotation)
- allows to change any Pod configuration since Pods are recreated
  - STS does not allow to change a lot of initial configuration
- target specific Pods via DNS via per-Pod created Service (e.g. redis-0 is a Service)
- since we deployed this via one-off kapp invocation if Pod gets deleted from the cluster, it will not be recreated
  - to address this, make sure that kapp deploy is continiously executed
    - simplest way to do that is to use [kapp-controller](https://github.com/k14s/kapp-controller)
  - (...it would be nice to create a tiny dedicated controller that can keep one named Pod alive)
