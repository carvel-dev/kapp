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
- allows to change any configuration in any way since all resources are managed directly
- target specific Pods via DNS via per-Pod created Service (e.g. redis-0 is a Service)
- since we deployed this via one-off kapp invocation if Pod gets deleted from the cluster, it will not be recreated
  - to address this, make sure that kapp deploy is continiously executed
    - simplest way to do that is to use [kapp-controller](https://github.com/vmware-tanzu/carvel-kapp-controller)
  - (...it would be nice to create a tiny dedicated controller that can keep one named Pod alive)

Note: this example uses Redis, but I actually did not bother configuring Redis replication between Pods since this was not the point of this example.

Deployed output:

```
$ kapp deploy -a redis -f <(ytt -f examples/sts-alternative)
Target cluster 'https://x.x.x.x' (nodes: gke-dk-jan-9-pool-1-d01bcd06-52ch, 4+)

Changes

Namespace  Name                Kind                   Conds.  Age  Op      Op st.  Wait to    Rs  Ri
default    redis               Service                -       -    create  -       reconcile  -   -
^          redis-0             PersistentVolumeClaim  -       -    create  -       reconcile  -   -
^          redis-0             Pod                    -       -    create  -       reconcile  -   -
^          redis-0             Service                -       -    create  -       reconcile  -   -
^          redis-1             PersistentVolumeClaim  -       -    create  -       reconcile  -   -
^          redis-1             Pod                    -       -    create  -       reconcile  -   -
^          redis-1             Service                -       -    create  -       reconcile  -   -
^          redis-2             PersistentVolumeClaim  -       -    create  -       reconcile  -   -
^          redis-2             Pod                    -       -    create  -       reconcile  -   -
^          redis-2             Service                -       -    create  -       reconcile  -   -
^          redis-config-ver-1  ConfigMap              -       -    create  -       reconcile  -   -

Op:      11 create, 0 delete, 0 update, 0 noop
Wait to: 11 reconcile, 0 delete, 0 noop

Continue? [yN]: y

8:35:46AM: ---- applying 4 changes [0/11 done] ----
8:35:46AM: create configmap/redis-config-ver-1 (v1) namespace: default
8:35:47AM: create persistentvolumeclaim/redis-2 (v1) namespace: default
8:35:47AM: create persistentvolumeclaim/redis-0 (v1) namespace: default
8:35:47AM: create persistentvolumeclaim/redis-1 (v1) namespace: default
8:35:47AM: ---- waiting on 4 changes [0/11 done] ----
8:35:47AM: ok: reconcile persistentvolumeclaim/redis-0 (v1) namespace: default
8:35:47AM: ok: reconcile persistentvolumeclaim/redis-1 (v1) namespace: default
8:35:47AM: ok: reconcile persistentvolumeclaim/redis-2 (v1) namespace: default
8:35:47AM: ok: reconcile configmap/redis-config-ver-1 (v1) namespace: default
8:35:47AM: ---- applying 5 changes [4/11 done] ----
8:35:47AM: create service/redis-2 (v1) namespace: default
8:35:47AM: create service/redis (v1) namespace: default
8:35:47AM: create service/redis-1 (v1) namespace: default
8:35:47AM: create service/redis-0 (v1) namespace: default
8:35:49AM: create pod/redis-0 (v1) namespace: default
8:35:49AM: ---- waiting on 5 changes [4/11 done] ----
8:35:49AM: ok: reconcile service/redis (v1) namespace: default
8:35:49AM: ok: reconcile service/redis-2 (v1) namespace: default
8:35:49AM: ok: reconcile service/redis-1 (v1) namespace: default
8:35:49AM: ok: reconcile service/redis-0 (v1) namespace: default
8:35:49AM: ongoing: reconcile pod/redis-0 (v1) namespace: default
8:35:49AM:  ^ Pending: ContainerCreating
8:35:49AM: ---- waiting on 1 changes [8/11 done] ----
8:35:59AM: ok: reconcile pod/redis-0 (v1) namespace: default
8:35:59AM: ---- applying 1 changes [9/11 done] ----
8:36:01AM: create pod/redis-1 (v1) namespace: default
8:36:01AM: ---- waiting on 1 changes [9/11 done] ----
8:36:01AM: ongoing: reconcile pod/redis-1 (v1) namespace: default
8:36:01AM:  ^ Pending: ContainerCreating
8:36:10AM: ok: reconcile pod/redis-1 (v1) namespace: default
8:36:10AM: ---- applying 1 changes [10/11 done] ----
8:36:12AM: create pod/redis-2 (v1) namespace: default
8:36:12AM: ---- waiting on 1 changes [10/11 done] ----
8:36:12AM: ongoing: reconcile pod/redis-2 (v1) namespace: default
8:36:12AM:  ^ Pending: ContainerCreating
8:36:29AM: ok: reconcile pod/redis-2 (v1) namespace: default
8:36:29AM: ---- applying complete [11/11 done] ----
8:36:29AM: ---- waiting complete [11/11 done] ----

Succeeded
```
