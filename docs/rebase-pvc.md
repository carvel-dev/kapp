## PersistentVolumeClaim rebase

Here is an example on how to use custom `rebaseRules` to "prefer" server chosen value for several annotations added by PVC controller (in other words, cluster owned fields), instead of removing them based on given configuration.

Let's deploy via `kapp deploy -a test -f config.yml -c` with following configuration `config.yml`:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysqlclaim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

Without additional rebase rules following diff will be presented upon next deploy, stating that several annotations will be removed (since they were not present in the initial configuration):

```bash
$ kapp deploy -a test -f config.yml -c

Target cluster 'https://x.x.x.x' (nodes: gke-dk-jan-9-default-pool-a218b1c9-55sl, 3+)

--- update persistentvolumeclaim/mysqlclaim (v1) namespace: default
  ...
  2,  2   metadata:
  3     -   annotations:
  4     -     pv.kubernetes.io/bind-completed: "yes"
  5     -     pv.kubernetes.io/bound-by-controller: "yes"
  6     -     volume.beta.kubernetes.io/storage-provisioner: kubernetes.io/gce-pd
  7,  3     creationTimestamp: "2020-03-08T22:17:29Z"
  8,  4     finalizers:
  ...
 24, 20     storageClassName: standard
 25     -   volumeMode: Filesystem
 26, 21     volumeName: pvc-1be63b2b-20de-429c-863a-9e7eb062f5d3
 27, 22   status:

Changes

Namespace  Name        Kind                   Conds.  Age  Op      Wait to    Rs  Ri
default    mysqlclaim  PersistentVolumeClaim  -       43s  update  reconcile  ok  -

Op:      0 create, 0 delete, 1 update, 0 noop
Wait to: 1 reconcile, 0 delete, 0 noop

Continue? [yN]:
```

To let kapp know that these annotations should be copied from the live resource copy, we can augment deploys with following configuration `kapp-config.yml`:

```yaml
---
apiVersion: kapp.k14s.io/v1alpha1
kind: Config

rebaseRules:
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

- path: [spec, volumeMode]
  type: copy
  sources: [new, existing]
  resourceMatchers: *pvcs
```

```bash
$ kapp deploy -a test -f config.yml -f rules.yml -c

Target cluster 'https://x.x.x.x' (nodes: gke-dk-jan-9-default-pool-a218b1c9-55sl, 3+)

Changes

Namespace  Name  Kind  Conds.  Age  Op  Wait to  Rs  Ri

Op:      0 create, 0 delete, 0 update, 0 noop
Wait to: 0 reconcile, 0 delete, 0 noop

Succeeded
```
