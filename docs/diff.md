## Diff

kapp compares resources specified in files against resources that exist in Kubernetes API. Once change set is calculated, it provides an option to apply it (see [Apply](apply.md) section for further details).

There are four different types of changes: `add`, `del`, `mod`, `keep`. They are visible in the `Changed` column of diff summary.

### Diff strategies

There are two diff strategies used by kapp:

1. kapp compares against last applied (previously by kapp) resource content **if** there were no outside changes done to the resource (done outside of kapp, for examples, by another team member); kapp tries to use this strategy as much as possible to produce more user-friendly diffs.

2. kapp compares against live resource **if** there were outside changes to the resource (hence, sometimes you may see a diff that shows several deleted fields even though these fields are not specified in the file)

Strategy is selected for each resource individually.

Related: [rebase rules](config.md).

### Ignored Changes

Some changes to resources are marked as ignored because associated resources are not managed (or created by) by kapp itself.

Common example is a Pod resource that is created by a deployments controller (provided by Kubernetes core) for a Deployment resource that was deployed by kapp. kapp knows about Pod resource because Deployment labels were propagated to the Pod (though not all controllers do that). kapp will not apply ignored changes (unless explicitly asked) because it assumes that controllers will manage them as their owner resource gets updated/deleted.

### Versioned Resources

In some cases it's useful to represent a change to a resource as a new resource. Common example is a workflow to update ConfigMap referenced by a Deployment. Deployments do not restart their Pods when ConfigMap changes making it tricky for wide variety of applications for pick up ConfigMap changes. kapp provides a solution for such scenarios, by offering a way to create uniquely named resources based on an original resource.

Anytime there is a change to a resource marked as a versioned resource, entirely new resource will be created instead of updating an existing resource. Additionally kapp follows configuration rules (default ones, and ones that can be provided as part of application) to find and update object references (since new resource name is not something that configuration author knew about).

To make resource versioned, add `kapp.k14s.io/versioned` annotation with an empty value. Created resource follow `{resource-name}-ver-{n}` naming pattern by incrementing `n` any time there is a change.

You can control number of kept resource versions via `kapp.k14s.io/num-versions=int` annotation. 

### Controlling diff via deploy flags

Diff summary shows quick information about what's being changed:

- `--diff-summary` (default `true`) shows diff summary, listing how resources have changed
- `--diff-summary-full` includes ignored and unchanged resources in diff summary

Diff changes (line-by-line diffs) are useful for looking at actual changes:

- `--diff-changes=bool` shows line-by-line diffs
- `--diff-changes-full=bool` includes ignored and unchanged resources in line-by-line diffs
- `--diff-context=int` controls number of lines to show around changed lines

Controlling how diffing is done:

- `--diff-against-last-applied=bool` forces kapp to use particular diffing strategy (see above)
- `--diff-run` stops after showing diff information
