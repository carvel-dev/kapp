## Diff

kapp compares resources specified in files against resources that exist in Kubernetes API. Once change set is calculated, it provides an option to apply it (see [Apply](apply.md) section for further details).

There are four different types of operations: `create`, `update`, `delete`, `noop` (shown as empty). Seen in `Op` column of diff summary table. Additionally there is `Op strategy` column (shorted as `Op st.`), added in v0.31.0+, that shows supplemental information how operation will be performed (for example `fallback on replace` for `update` operation).

There are three different types of waiting: `reconcile` (waits until resource has converged to its desired state; see [apply waiting](apply-waiting.md) for waiting semantics), `delete` (waits until resource is gone), `noop` (shown as empty). Seen in `Wait to` column of diff summary table.

### Diff strategies

There are two diff strategies used by kapp:

1. kapp compares against last applied resource content (previously applied by kapp; stored in annotation `kapp.k14s.io/original`) **if** there were no outside changes done to the resource (i.e. done outside of kapp, for example, by another team member or controller); kapp tries to use this strategy as much as possible to produce more user-friendly diffs.

2. kapp compares against live resource content **if** it detects there were outside changes to the resource (hence, sometimes you may see a diff that shows several deleted fields even though these fields are not specified in the original file)

Strategy is selected for each resource individually. You can control which strategy is used for all resources via `--diff-against-last-applied=bool` flag.

Related: [rebase rules](config.md).

### Versioned Resources

In some cases it's useful to represent an update to a resource as an entirely new resource. Common example is a workflow to update ConfigMap referenced by a Deployment. Deployments do not restart their Pods when ConfigMap changes making it tricky for wide variety of applications for pick up ConfigMap changes. kapp provides a solution for such scenarios, by offering a way to create uniquely named resources based on an original resource.

Anytime there is a change to a resource marked as a versioned resource, entirely new resource will be created instead of updating an existing resource. Additionally kapp follows configuration rules (default ones, and ones that can be provided as part of application) to find and update object references (since new resource name is not something that configuration author knew about).

To make resource versioned, add `kapp.k14s.io/versioned` annotation with an empty value. Created resource follow `{resource-name}-ver-{n}` naming pattern by incrementing `n` any time there is a change.

You can control number of kept resource versions via `kapp.k14s.io/num-versions=int` annotation.

Try deploying [redis-with-configmap example](../examples/gitops/redis-with-configmap) and changing `ConfigMap` in a next deploy.

### Controlling diff via resource annotations

#### kapp.k14s.io/disable-original

`kapp.k14s.io/disable-original` annotation controls whether to record provided resource copy (rarely wanted)

Possible values ``. In some cases it's not possible or wanted to record applied resource copy into its annotation `kapp.k14s.io/original`. One such case might be when resource is extremely lengthy (e.g. long ConfigMap or CustomResourceDefinition) and will exceed annotation value max length of 262144 bytes.

---
### Controlling diff via deploy flags

Diff summary shows quick information about what's being changed:

- `--diff-summary=bool` (default `true`) shows diff summary, listing how resources have changed

Diff changes (line-by-line diffs) are useful for looking at actual changes:

- `--diff-changes=bool` (`-c`) (default `false`) shows line-by-line diffs
- `--diff-context=int` (default `2`) controls number of lines to show around changed lines
- `--diff-mask=bool` (default `true`) controls whether to mask sensitive fields

Controlling how diffing is done:

- `--diff-against-last-applied=bool` (default `false`) forces kapp to use particular diffing strategy (see above)
- `--diff-run=bool` (default `false`) stops after showing diff information
- `--diff-exit-status=bool` (default `false`) controls exit status for diff runs (`0`: unused, `1`: any error, `2`: no changes, `3`: pending changes)
