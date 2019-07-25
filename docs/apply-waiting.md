### Apply Waiting

kapp has builtin knowledge on how to wait for the following resource types:

- any resource with `metadata.deletionTimestamp`: wait for resource to be fully removed
- any resource with reconciling annotations: [see below](#reconciling-annotations)
- `apiextensions.k8s.io/<any>/CustomResourceDefinition`: wait for all conditions to turn `True`
- `apps/v1/DaemonSet`: wait for `status.numberUnavailable` to be 0
- `apps/v1/Deployment`: see below
- `apps/v1/ReplicaSet`: wait for `status.replicas == status.availableReplicas`
- `batch/v1/Job`: wait for `Complete` or `Failed` conditions to appear
- `batch/<any>/CronJob`: immediately considers as done
- `/v1/Pod`: looks at `status.phase`
- `/v1/Service`: wait for `spec.clusterIP` and/or `status.loadBalancer.ingress` to become set

#### Deployment resource

kapp by default waits for `apps/v1/Deployment` resource to have `status.unavailableReplicas` equal to zero. Additionally waiting behaviour can be controlled via following annotations:

- `kapp.k14s.io/apps-v1-deployment-wait-minimum-replicas-available` annotation controls how many new available replicas are enough to consider waiting successful. Example values: `"10"`, `"5%"`.

#### Reconciling annotations

If resource has below annotations, kapp will wait for reconcile state to become either `ok` or `fail`. Typically these annotations will be controlled by a controller/operator running in cluster.

- `kapp.k14s.io/reconcile-state`: indicates current reconcilation state. Possible values: `ok`, `fail`, `ongoing`.
- `kapp.k14s.io/reconcile-info`: includes additional information about current reconcilation state.
