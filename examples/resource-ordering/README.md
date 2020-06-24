# Resource Ordering

In this example we are going to deploy Redis primary and replica, and after both deployed, run a job to check if they are being synced.

Note in the example output below you can see that `deployment/redis-primary`, then `deployment/redis-replica` and finally `job/sync-check` are deployed:

```
$ kapp deploy -a redis -f examples/resource-ordering/
Changes

Namespace  Name           Kind        Conds.  Age  Op      Wait to
default    redis-primary  Deployment  -       -    create  reconcile
~          redis-primary  Service     -       -    create  reconcile
~          redis-replica  Deployment  -       -    create  reconcile
~          redis-replica  Service     -       -    create  reconcile
~          sync-check     Job         -       -    create  reconcile

Op:      5 create, 0 delete, 0 update, 0 noop
Wait to: 5 reconcile, 0 delete, 0 noop

Continue? [yN]: y

5:05:02PM: ---- applying 3 changes [0/5 done] ----
5:05:02PM: create service/redis-primary (v1) namespace: default
5:05:02PM: create deployment/redis-primary (apps/v1) namespace: default
5:05:02PM: create service/redis-replica (v1) namespace: default
5:05:02PM: ---- waiting on 3 changes [0/5 done] ----
5:05:03PM: ok: reconcile service/redis-primary (v1) namespace: default
5:05:03PM: ongoing: reconcile deployment/redis-primary (apps/v1) namespace: default
5:05:03PM:  ^ Waiting for 1 unavailable replicas
5:05:03PM:  L ok: waiting on replicaset/redis-primary-7dc969599b (apps/v1) namespace: default
5:05:03PM:  L ongoing: waiting on pod/redis-primary-7dc969599b-4lz22 (v1) namespace: default
5:05:03PM:     ^ Pending: ContainerCreating
5:05:04PM: ok: reconcile service/redis-replica (v1) namespace: default
5:05:04PM: ---- waiting on 1 changes [2/5 done] ----
5:05:04PM: ongoing: reconcile deployment/redis-primary (apps/v1) namespace: default
5:05:04PM:  ^ Waiting for 1 unavailable replicas
5:05:04PM:  L ok: waiting on replicaset/redis-primary-7dc969599b (apps/v1) namespace: default
5:05:04PM:  L ok: waiting on podmetrics/redis-primary-7dc969599b-4lz22 (metrics.k8s.io/v1beta1) namespace: default
5:05:04PM:  L ongoing: waiting on pod/redis-primary-7dc969599b-4lz22 (v1) namespace: default
5:05:04PM:     ^ Pending: ContainerCreating
5:05:10PM: ongoing: reconcile deployment/redis-primary (apps/v1) namespace: default
5:05:10PM:  ^ Waiting for 1 unavailable replicas
5:05:10PM:  L ok: waiting on replicaset/redis-primary-7dc969599b (apps/v1) namespace: default
5:05:10PM:  L ok: waiting on podmetrics/redis-primary-7dc969599b-4lz22 (metrics.k8s.io/v1beta1) namespace: default
5:05:10PM:  L ongoing: waiting on pod/redis-primary-7dc969599b-4lz22 (v1) namespace: default
5:05:10PM:     ^ Condition Ready is not True (False)
5:05:11PM: ok: reconcile deployment/redis-primary (apps/v1) namespace: default
5:05:11PM: ---- applying 1 changes [3/5 done] ----
5:05:11PM: create deployment/redis-replica (apps/v1) namespace: default
5:05:13PM: ---- waiting on 1 changes [3/5 done] ----
5:05:13PM: ongoing: reconcile deployment/redis-replica (apps/v1) namespace: default
5:05:13PM:  ^ Waiting for generation 2 to be observed
5:05:13PM:  L ok: waiting on replicaset/redis-replica-dd49d97bc (apps/v1) namespace: default
5:05:13PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-mqf2t (v1) namespace: default
5:05:13PM:     ^ Pending: ContainerCreating
5:05:13PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-9hnss (v1) namespace: default
5:05:13PM:     ^ Condition Ready is not True (False)
5:05:15PM: ongoing: reconcile deployment/redis-replica (apps/v1) namespace: default
5:05:15PM:  ^ Waiting for 2 unavailable replicas
5:05:15PM:  L ok: waiting on replicaset/redis-replica-dd49d97bc (apps/v1) namespace: default
5:05:15PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-mqf2t (v1) namespace: default
5:05:15PM:     ^ Pending: ContainerCreating
5:05:15PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-9hnss (v1) namespace: default
5:05:15PM:     ^ Condition Ready is not True (False)
5:05:18PM: ongoing: reconcile deployment/redis-replica (apps/v1) namespace: default
5:05:18PM:  ^ Waiting for 2 unavailable replicas
5:05:18PM:  L ok: waiting on replicaset/redis-replica-dd49d97bc (apps/v1) namespace: default
5:05:18PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-mqf2t (v1) namespace: default
5:05:18PM:     ^ Pending: ContainerCreating
5:05:18PM:  L ok: waiting on pod/redis-replica-dd49d97bc-9hnss (v1) namespace: default
5:05:19PM: ongoing: reconcile deployment/redis-replica (apps/v1) namespace: default
5:05:19PM:  ^ Waiting for 1 unavailable replicas
5:05:19PM:  L ok: waiting on replicaset/redis-replica-dd49d97bc (apps/v1) namespace: default
5:05:19PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-mqf2t (v1) namespace: default
5:05:19PM:     ^ Pending: ContainerCreating
5:05:19PM:  L ok: waiting on pod/redis-replica-dd49d97bc-9hnss (v1) namespace: default
5:05:20PM: ongoing: reconcile deployment/redis-replica (apps/v1) namespace: default
5:05:20PM:  ^ Waiting for 1 unavailable replicas
5:05:20PM:  L ok: waiting on replicaset/redis-replica-dd49d97bc (apps/v1) namespace: default
5:05:20PM:  L ongoing: waiting on pod/redis-replica-dd49d97bc-mqf2t (v1) namespace: default
5:05:20PM:     ^ Condition Ready is not True (False)
5:05:20PM:  L ok: waiting on pod/redis-replica-dd49d97bc-9hnss (v1) namespace: default
5:05:29PM: ok: reconcile deployment/redis-replica (apps/v1) namespace: default
5:05:29PM: ---- applying 1 changes [4/5 done] ----
5:05:29PM: create job/sync-check (batch/v1) namespace: default
5:05:29PM: ---- waiting on 1 changes [4/5 done] ----
logs | # waiting for 'sync-check-xvspr > sync-check' logs to become available...
5:05:30PM: ongoing: reconcile job/sync-check (batch/v1) namespace: default
5:05:30PM:  ^ Waiting to complete (1 active, 0 failed, 0 succeeded)
5:05:30PM:  L ongoing: waiting on pod/sync-check-xvspr (v1) namespace: default
5:05:30PM:     ^ Pending: ContainerCreating
5:05:31PM: ongoing: reconcile job/sync-check (batch/v1) namespace: default
5:05:31PM:  ^ Waiting to complete (1 active, 0 failed, 0 succeeded)
5:05:31PM:  L ok: waiting on pod/sync-check-xvspr (v1) namespace: default
logs | # starting tailing 'sync-check-xvspr > sync-check' logs
logs | sync-check-xvspr > sync-check | + primary=-h redis-primary -p 6379
logs | sync-check-xvspr > sync-check | + replica=-h redis-replica -p 6379
logs | sync-check-xvspr > sync-check | + redis-cli -h redis-primary -p 6379 ping
logs | sync-check-xvspr > sync-check | PONG
logs | sync-check-xvspr > sync-check | + redis-cli -h redis-replica -p 6379 ping
logs | sync-check-xvspr > sync-check | PONG
logs | sync-check-xvspr > sync-check | + redis-cli -h redis-replica -p 6379 info
logs | sync-check-xvspr > sync-check | + grep -A10 Replication
logs | sync-check-xvspr > sync-check | + redis-cli -h redis-primary -p 6379 set key val
logs | sync-check-xvspr > sync-check | # Replication
logs | sync-check-xvspr > sync-check | role:slave
logs | sync-check-xvspr > sync-check | master_host:redis-primary
logs | sync-check-xvspr > sync-check | master_port:6379
logs | sync-check-xvspr > sync-check | master_link_status:up
logs | sync-check-xvspr > sync-check | master_last_io_seconds_ago:2
logs | sync-check-xvspr > sync-check | master_sync_in_progress:0
logs | sync-check-xvspr > sync-check | slave_repl_offset:28
logs | sync-check-xvspr > sync-check | slave_priority:100
logs | sync-check-xvspr > sync-check | slave_read_only:1
logs | sync-check-xvspr > sync-check | connected_slaves:0
logs | sync-check-xvspr > sync-check | + sleep 2
logs | sync-check-xvspr > sync-check | OK
logs | sync-check-xvspr > sync-check | + redis-cli -h redis-replica -p 6379 get key
logs | sync-check-xvspr > sync-check | + redis-cli -h redis-replica -p 6379 get key
logs | sync-check-xvspr > sync-check | val
logs | sync-check-xvspr > sync-check | + result=val
logs | sync-check-xvspr > sync-check | + [ xval != xval ]
5:05:34PM: ok: reconcile job/sync-check (batch/v1) namespace: default
5:05:34PM: ---- applying complete [5/5 done] ----
5:05:34PM: ---- waiting complete [5/5 done] ----

Succeeded
```
