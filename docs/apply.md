## Apply

Once change set is calculated (see [Diff](diff.md) section for details), kapp asks for user confirmation (unless `--yes` flag is specified) to proceed with changes.

Changes are applied in particular order. Currently it is:

1. CRDs creation/update
1. Namespace creation/update
1. other resources
1. CRD deletion

All created resources are labeled with several labels:

- `kapp.k14s.io/app` to track which application "owns" resource
- `kapp.k14s.io/identity` to identify preferred API version used when creating resource
- `kapp.k14s.io/association` to track (best effort) parent-child relationships between resources

Every time application is deployed, new application change record is saved. They can be viewed via `kapp app-change ls -a app-name`.

Related: [ownership label rules](config.md) and [label scoping rules](config.md).

### Controlling apply via resource annotations

- `kapp.k14s.io/update-strategy` annotation controls update behaviour

	Possible values: `` (default), `fallback-on-replace`, `always-replace`. In some cases entire resources or subset resource fields are immutable which forces kapp users to specify how to apply wanted update.

	- `` means to issue plain update call
	- `fallback-on-replace` causes kapp to fallback to resource replacement if update call results in `Invalid` error
	- `always-replace` causes kapp to always delete and then create resource

- `kapp.k14s.io/delete-strategy` annotation controls deletion behaviour

	Possible values: `` (default), `orphan`. By default resource is deleted, however; choosing `orphan` value will make kapp forget about this resource. Note that if this resource is owned by a different resource that's being deleted, it might still get deleted. Orphaned resources are annotated with `kapp.k14s.io/orphaned` annotation.

### Controlling apply via deploy flags

- `--apply-ignored=bool` explicitly applies ignored changes; this is useful in cases when controllers lose track of some resources instead of for example deleting them
- `--apply-wait=bool` (default `true`) controls whether kapp will wait for resource to "stabilize"
- `--apply-wait-ignored=bool` controls whether kapp will wait for ignored changes (regardless whether they were initiated by kapp or by controllers)
- `--apply-default-update-strategy=string` controls default strategy for all resources (see `kapp.k14s.io/update-strategy` annotation above)
