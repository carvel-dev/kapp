## Apply

Once change set is calculated (see [Diff](diff.md) section for details), kapp asks for user confirmation (unless `--yes` flag is specified) to proceed with changes.

Changes are applied in particular order as described in [Apply ordering](apply-ordering.md).

All created resources are labeled with several labels:

- `kapp.k14s.io/app` to track which application "owns" resource
- `kapp.k14s.io/identity` to identify preferred API version used when creating resource
- `kapp.k14s.io/association` to track (best effort) parent-child relationships between resources

Every time application is deployed, new application change record is saved. They can be viewed via `kapp app-change ls -a app-name`.

Related: [ownership label rules](config.md) and [label scoping rules](config.md).

### Controlling apply via resource annotations

#### kapp.k14s.io/create-strategy

`kapp.k14s.io/create-strategy` annotation controls create behaviour (rarely necessary)

Possible values `` (default), `fallback-on-update`. In some cases creation of a resource may conflict with that resource being created in the cluster by other means (often automated). An example of that is creation of default ServiceAccount by kapp racing with Kubernetes service accounts controller doing the same thing. By specifying `fallback-on-update` value, kapp will catch resource creation conflicts and apply resource as an update.

#### kapp.k14s.io/update-strategy

`kapp.k14s.io/update-strategy` annotation controls update behaviour

Possible values: `` (default), `fallback-on-replace`, `always-replace`, `skip`. In some cases entire resources or subset resource fields are immutable which forces kapp users to specify how to apply wanted update.

- `` means to issue plain update call
- `fallback-on-replace` causes kapp to fallback to resource replacement if update call results in `Invalid` error. Note that if resource is replaced (= delete + create), it may be negatively affected (loss of persistent data, loss of availability, etc.). For example, if Deployment or DaemonSet is first deleted and then created then associated Pods will be recreated as well, but all at the same time (even if rolling update is enabled), which likely causes an availability gap.
- `always-replace` causes kapp to always delete and then create resource (See note above as well.)
- `skip` causes kapp to not apply update (it will show up in a diff next time). Available in v0.33.0+.

#### kapp.k14s.io/delete-strategy

`kapp.k14s.io/delete-strategy` annotation controls deletion behaviour

Possible values: `` (default), `orphan`. By default resource is deleted, however; choosing `orphan` value will make kapp forget about this resource. Note that if this resource is owned by a different resource that's being deleted, it might still get deleted. Orphaned resources are labeled with `kapp.k14s.io/orphaned` label. As of v0.31.0+, resource is also disassociated from owning app so that it can be owned by future apps.

#### kapp.k14s.io/owned-for-deletion

`kapp.k14s.io/owned-for-deletion` annotation controls resource deletion during `kapp delete` command

Possible values: ``. By default non-kapp owned resources are not explicitly deleted by kapp, but expected to be deleted by the cluster (for example Endpoints resource for each Service). In some cases it's desired to annotate non-kapp owned resource so that it does get explicitly deleted, possibly because cluster does not plan to delete it (e.g. PVCs created by StatefulSet are not deleted by StatefulSet controller; [https://github.com/k14s/kapp/issues/36](https://github.com/k14s/kapp/issues/36)).

#### kapp.k14s.io/nonce

`kapp.k14s.io/nonce` annotation allows to inject unique ID

Possible values: `` (default). Annotation value will be replaced with a unique ID on each deploy. This allows to force resource update as value changes every time.

#### kapp.k14s.io/deploy-logs

`kapp.k14s.io/deploy-logs` annotation indicates which Pods' log output to show during deploy

Possible values: `` (default; equivalent to `for-new`), `for-new` (only newly created Pods are tailed), `for-existing` (only existing Pods are tailed), `for-new-or-existing` (both newly created and existing Pods are tailed). Especially useful when added to Jobs. For example, see [examples/resource-ordering/sync-check.yml](../examples/resource-ordering/sync-check.yml)

#### kapp.k14s.io/deploy-logs-container-names

`kapp.k14s.io/deploy-logs-container-names` annotation indicates which Containers' log output to show during deploy

Possible values: `` (default), 'containerName1', 'containerName1,containerName2'

---
### Controlling apply via deploy flags

- `--apply-ignored=bool` explicitly applies ignored changes; this is useful in cases when controllers lose track of some resources instead of for example deleting them
- `--apply-default-update-strategy=string` controls default strategy for all resources (see `kapp.k14s.io/update-strategy` annotation above)
- `--apply-exit-status=bool` (default `false`) controls exit status (`0`: unused, `1`: any error, `2`: no changes applied, `3`: at least one change applied)
- `--wait=bool` (default `true`) controls whether kapp will wait for resource to "stabilize". See [Apply waiting](apply-waiting.md)
- `--wait-ignored=bool` controls whether kapp will wait for ignored changes (regardless whether they were initiated by kapp or by controllers)
- `--logs=bool` (default `true`) controls whether to show logs as part of deploy output for Pods annotated with `kapp.k14s.io/deploy-logs: ""`
- `--logs-all=bool` (deafult `false`) controls whether to show all logs as part of deploy output for all Pods
