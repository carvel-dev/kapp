### Apply Waiting

kapp includes builtin rules on how to wait for the following resource types:

- [any resource with `metadata.deletionTimestamp`](../pkg/kapp/resourcesmisc/deleting.go): wait for resource to be fully removed
- [any resource with `kapp.k14s.io/reconcile-*` annotations](../pkg/kapp/resourcesmisc/reconciling.go): [see "Custom waiting behaviour" below](#custom-waiting-behaviour)
- [`apiextensions.k8s.io/<any>/CustomResourceDefinition`](../pkg/kapp/resourcesmisc/api_extensions_vx_crd.go): wait for all conditions to turn `True`
- [`apps/v1/DaemonSet`](../pkg/kapp/resourcesmisc/apps_v1_daemon_set.go): wait for `status.numberUnavailable` to be 0
- [`apps/v1/Deployment`](../pkg/kapp/resourcesmisc/apps_v1_deployment.go): [see "apps/v1/Deployment resource" below](#apps-v1-deployment-resource)
- [`apps/v1/ReplicaSet`](../pkg/kapp/resourcesmisc/apps_v1_replica_set.go): wait for `status.replicas == status.availableReplicas`
- [`batch/v1/Job`](../pkg/kapp/resourcesmisc/batch_v1_job.go): wait for `Complete` or `Failed` conditions to appear
- [`batch/<any>/CronJob`](../pkg/kapp/resourcesmisc/batch_vx_cron_job.go): immediately considers as done
- [`/v1/Pod`](../pkg/kapp/resourcesmisc/core_v1_pod.go): looks at `status.phase`
- [`/v1/Service`](../pkg/kapp/resourcesmisc/core_v1_service.go): wait for `spec.clusterIP` and/or `status.loadBalancer.ingress` to become set

If resource is not affected by the above rules, its waiting behaviour depends on aggregate of waiting states of its associated resources (associated resources are resources that share same `kapp.k14s.io/association` label value).

#### Controlling waiting via resource annotations

- `kapp.k14s.io/disable-wait` annotation controls whether waiting will happen at all. Possible values: ``.
- `kapp.k14s.io/disable-associated-resources-wait` annotation controls whether associated resources impact resource's waiting state. Possible values: ``.

#### apps/v1/Deployment resource

kapp by default waits for `apps/v1/Deployment` resource to have `status.unavailableReplicas` equal to zero. Additionally waiting behaviour can be controlled via following annotations:

- `kapp.k14s.io/apps-v1-deployment-wait-minimum-replicas-available` annotation controls how many new available replicas are enough to consider waiting successful. Example values: `"10"`, `"5%"`.

#### Custom waiting behaviour

kapp can be extended with custom waiting behaviour through resource annotations. Controllers/operators can update resource annotations to indicate resource's reconcilation state.

If resource has below annotations, kapp will wait for reconcile state to become either `ok` or `fail`.

- `kapp.k14s.io/reconcile-state`: indicates current reconcilation state. Possible values: `ok`, `fail`, `ongoing`.
- `kapp.k14s.io/reconcile-info`: includes additional information about current reconcilation state.

Note that it's recommended to have a mutating webhook hook to reset `kapp.k14s.io/reconcile-state` annotation to `ongoing` upon resource creation or update to avoid race between kapp and controller updating annotations.
