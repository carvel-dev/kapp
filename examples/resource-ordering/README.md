# Resource Ordering

In this example we are going to deploy Redis master and slave, and after both deployed, run a job to check if they are being synced.

Note in the example output below you can see that `deployment/redis-master`, then `deployment/redis-slave` and finally `job/sync-check` are deployed:

```
argonaut:kapp argonaut$ kapp deploy -a redis -f examples/resource-ordering/ -y
Changes

Namespace  Name          Kind        Conds.  Age  Op      Wait to
default    redis-master  Deployment  -       -    create  reconcile
~          redis-master  Service     -       -    create  reconcile
~          redis-slave   Deployment  -       -    create  reconcile
~          redis-slave   Service     -       -    create  reconcile
~          sync-check    Job         -       -    create  reconcile

5 create, 0 delete, 0 update

5 changes

3:29:06PM: ---- applying 3 changes [0/5 done] ----
3:29:06PM: create service/redis-master (v1) namespace: default
3:29:06PM: create deployment/redis-master (apps/v1) namespace: default
3:29:06PM: create service/redis-slave (v1) namespace: default

3:29:06PM: ---- waiting on 3 changes [0/5 done] ----
3:29:06PM: waiting on reconcile service/redis-master (v1) namespace: default
3:29:06PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default
3:29:07PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:07PM:  L waiting on replicaset/redis-master-5c77df79b4 (extensions/v1beta1) namespace: default ... done
3:29:07PM:  L waiting on pod/redis-master-5c77df79b4-zh88n (v1) namespace: default ... in progress: Pending: ContainerCreating
3:29:07PM: waiting on reconcile service/redis-slave (v1) namespace: default

3:29:07PM: ---- waiting on 1 changes [2/5 done] ----
3:29:07PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default
3:29:07PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:07PM:  L waiting on replicaset/redis-master-5c77df79b4 (extensions/v1beta1) namespace: default ... done
3:29:07PM:  L waiting on pod/redis-master-5c77df79b4-zh88n (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:08PM: ---- waiting on 1 changes [2/5 done] ----
3:29:08PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default
3:29:08PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:08PM:  L waiting on replicaset/redis-master-5c77df79b4 (apps/v1) namespace: default ... done
3:29:08PM:  L waiting on pod/redis-master-5c77df79b4-zh88n (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:09PM: ---- waiting on 1 changes [2/5 done] ----
3:29:09PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default
3:29:09PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:09PM:  L waiting on replicaset/redis-master-5c77df79b4 (extensions/v1beta1) namespace: default ... done
3:29:09PM:  L waiting on pod/redis-master-5c77df79b4-zh88n (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:10PM: ---- waiting on 1 changes [2/5 done] ----
3:29:10PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default
3:29:10PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:10PM:  L waiting on replicaset/redis-master-5c77df79b4 (apps/v1) namespace: default ... done
3:29:10PM:  L waiting on pod/redis-master-5c77df79b4-zh88n (v1) namespace: default ... in progress: Condition Ready is not True (False)

3:29:11PM: ---- waiting on 1 changes [2/5 done] ----
3:29:11PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default
3:29:11PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:11PM:  L waiting on replicaset/redis-master-5c77df79b4 (apps/v1) namespace: default ... done
3:29:11PM:  L waiting on pod/redis-master-5c77df79b4-zh88n (v1) namespace: default ... in progress: Condition Ready is not True (False)

3:29:12PM: ---- waiting on 1 changes [2/5 done] ----
3:29:12PM: waiting on reconcile deployment/redis-master (apps/v1) namespace: default

3:29:13PM: ---- applying 1 changes [3/5 done] ----
3:29:13PM: create deployment/redis-slave (apps/v1) namespace: default

3:29:13PM: ---- waiting on 1 changes [3/5 done] ----
3:29:13PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:13PM:  ^  ... in progress: Waiting for generation 2 to be observed
3:29:13PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1beta2) namespace: default ... done
3:29:13PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Pending
3:29:13PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:14PM: ---- waiting on 1 changes [3/5 done] ----
3:29:14PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:14PM:  ^  ... in progress: Waiting for 2 unavailable replicas
3:29:14PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1) namespace: default ... done
3:29:14PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Pending: ContainerCreating
3:29:14PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:15PM: ---- waiting on 1 changes [3/5 done] ----
3:29:15PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:15PM:  ^  ... in progress: Waiting for 2 unavailable replicas
3:29:15PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (extensions/v1beta1) namespace: default ... done
3:29:15PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Pending: ContainerCreating
3:29:15PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:16PM: ---- waiting on 1 changes [3/5 done] ----
3:29:16PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:16PM:  ^  ... in progress: Waiting for 2 unavailable replicas
3:29:16PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1) namespace: default ... done
3:29:16PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Pending: ContainerCreating
3:29:16PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... in progress: Condition Ready is not True (False)

3:29:17PM: ---- waiting on 1 changes [3/5 done] ----
3:29:17PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:17PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:17PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (extensions/v1beta1) namespace: default ... done
3:29:17PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Pending: ContainerCreating
3:29:17PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... done

3:29:18PM: ---- waiting on 1 changes [3/5 done] ----
3:29:18PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:19PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:19PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1beta2) namespace: default ... done
3:29:19PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Condition Ready is not True (False)
3:29:19PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... done

3:29:20PM: ---- waiting on 1 changes [3/5 done] ----
3:29:20PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:20PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:20PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1beta2) namespace: default ... done
3:29:20PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Condition Ready is not True (False)
3:29:20PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... done

3:29:21PM: ---- waiting on 1 changes [3/5 done] ----
3:29:21PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:21PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:21PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (extensions/v1beta1) namespace: default ... done
3:29:21PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Condition Ready is not True (False)
3:29:21PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... done

3:29:22PM: ---- waiting on 1 changes [3/5 done] ----
3:29:22PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:22PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:22PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1beta2) namespace: default ... done
3:29:22PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Condition Ready is not True (False)
3:29:22PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... done

3:29:23PM: ---- waiting on 1 changes [3/5 done] ----
3:29:23PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default
3:29:23PM:  ^  ... in progress: Waiting for 1 unavailable replicas
3:29:23PM:  L waiting on replicaset/redis-slave-67cfc7c9fc (apps/v1) namespace: default ... done
3:29:23PM:  L waiting on pod/redis-slave-67cfc7c9fc-lnqlw (v1) namespace: default ... in progress: Condition Ready is not True (False)
3:29:23PM:  L waiting on pod/redis-slave-67cfc7c9fc-hm647 (v1) namespace: default ... done

3:29:24PM: ---- waiting on 1 changes [3/5 done] ----
3:29:24PM: waiting on reconcile deployment/redis-slave (apps/v1) namespace: default

3:29:24PM: ---- applying 1 changes [4/5 done] ----
3:29:24PM: create job/sync-check (batch/v1) namespace: default

3:29:24PM: ---- waiting on 1 changes [4/5 done] ----
3:29:24PM: waiting on reconcile job/sync-check (batch/v1) namespace: default
3:29:24PM:  ^  ... in progress: Waiting to complete (0 active, 0 failed, 0 succeeded)
3:29:24PM:  L waiting on pod/sync-check-j9rkj (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:25PM: ---- waiting on 1 changes [4/5 done] ----
3:29:25PM: waiting on reconcile job/sync-check (batch/v1) namespace: default
3:29:26PM:  ^  ... in progress: Waiting to complete (1 active, 0 failed, 0 succeeded)
3:29:26PM:  L waiting on pod/sync-check-j9rkj (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:27PM: ---- waiting on 1 changes [4/5 done] ----
3:29:27PM: waiting on reconcile job/sync-check (batch/v1) namespace: default
3:29:27PM:  ^  ... in progress: Waiting to complete (1 active, 0 failed, 0 succeeded)
3:29:27PM:  L waiting on pod/sync-check-j9rkj (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:28PM: ---- waiting on 1 changes [4/5 done] ----
3:29:28PM: waiting on reconcile job/sync-check (batch/v1) namespace: default
3:29:28PM:  ^  ... in progress: Waiting to complete (1 active, 0 failed, 0 succeeded)
3:29:28PM:  L waiting on pod/sync-check-j9rkj (v1) namespace: default ... in progress: Pending: ContainerCreating

3:29:29PM: ---- waiting on 1 changes [4/5 done] ----
3:29:29PM: waiting on reconcile job/sync-check (batch/v1) namespace: default
3:29:29PM:  ^  ... in progress: Waiting to complete (1 active, 0 failed, 0 succeeded)
3:29:29PM:  L waiting on pod/sync-check-j9rkj (v1) namespace: default ... done

3:29:30PM: ---- waiting on 1 changes [4/5 done] ----
3:29:30PM: waiting on reconcile job/sync-check (batch/v1) namespace: default

3:29:30PM: ---- changes applied ----

Succeeded
```
