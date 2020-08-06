## FAQ

### `Error: Asking for confirmation: EOF`

This probably means you have piped configuration into kapp and did not specify `--yes` (`-y`) flag to continue. It's necessary because kapp can no longer ask for confirmation via stdin. Feel free to re-run the command with `--diff-changes` (`-c`) to make sure pending changes are correct. Instead of using a pipe you can also use an anonymous fifo keeping stdin free for the confirmation prompt, e.g. `kapp deploy -a app1 -f <(ytt -f config/)`

---
### Where to store app resources (i.e. in which namespace)?

See [state namespace](state-namespace.md) doc page.

---
### `... Field is immutable` error

> After changing the labels/selectors in one of my templates, I'm getting the `MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable (reason: Invalid)` errors on deployment resource. Is there a way to tell kapp to force the change?

[via slack](https://kubernetes.slack.com/archives/CH8KCCKA5/p1565600090224400)

Some fields on a resource are immutable. kapp provides a `kapp.k14s.io/update-strategy` annotation that controls how kapp will update resource. One of the strategies is `fallback-on-replace` which will have kapp recreate an object (delete, wait, then create) if initial update results in `Invalid` error. See [Controlling apply via resource annotations](https://github.com/k14s/kapp/blob/develop/docs/apply.md#controlling-apply-via-resource-annotations) for details.

---
### `Job.batch is invalid: ... spec.selector: Required value` error

`batch.Job` resource is augmented by the Job controller with unique labels upon its creation. When using kapp to subsequently update existing Job resource, API server will return `Invalid` error since given configuration does not include `spec.selector`, and `job-name` and `controller-uid` labels. kapp's [rebase rules](https://github.com/k14s/kapp/blob/develop/docs/config.md#rebaserules) can be used to copy over necessary configuration from server side copy; however, since Job resource is mostly immutable, we recommend to use [`kapp.k14s.io/update-strategy` annotation](https://github.com/k14s/kapp/blob/develop/docs/apply.md#kappk14sioupdate-strategy) set to `fallback-on-replace` to recreate Job resource with any updates.

---
### Updating Deployments when ConfigMap changes

> Can kapp force update on ConfigMaps in Deployments/DaemonSets? Just noticed that it didn't do that and I somehow expected it to.

[via slack](https://kubernetes.slack.com/archives/CH8KCCKA5/p1565624685226400)

kapp has a feature called [versioned resources](diff.md#versioned-resources) that allows kapp to create uniquely named resources instead of updating resources with changes. Resources referencing versioned resources are forced to be updated with new names, and therefore are changed, thus solving a problem of how to propagate changes safely.

---
### Quick way to find common kapp command variations

See [cheatsheet](cheatsheet.md).

---
### Limit number of ReplicaSets for Deployments

> Everytime I do a new deploy w/ kapp I see a new replicaset, along with all of the previous ones.

[via slack](https://kubernetes.slack.com/archives/CH8KCCKA5/p1565887856281400)

`Deployment` resource has a field `.spec.revisionHistoryLimit` that controls how many previous `ReplicaSets` to keep. See [Deployment's clean up polciy](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#clean-up-policy) for more details.

---
### Changes detected immediately after successful deploy

Sometimes Kubernetes API server will convert submitted field values into their canonical form server-side. This will be detected by kapp as a change during a next deploy. To avoid such changes in future, you will have to change your provided field values to what API server considers as canonical.

```
...
186     -  cpu: "2"
187     -  memory: 1Gi
    170 +  cpu: 2000m
    171 +  memory: 1024Mi
...
```

Consider using [ytt](https://get-ytt.io) and [its overlay feature](https://github.com/k14s/ytt/blob/develop/docs/lang-ref-ytt-overlay.md) to change values if you do not control source configuration.

---
### Changes detected after resource is modified server-side

There might be cases where other system actors (various controllers) may modify resource outside of kapp. Common example is Deployment's `spec.replicas` field is modified by Horizontal Pod Autoscaler controller. To let kapp know of such external behaviour use custom `rebaseRules` configuration (see [HPA and Deployment rebase](https://github.com/k14s/kapp/blob/develop/docs/hpa-deployment-rebase.md) for details).

---
### Colors are not showing up in my CI build, in my terminal, etc.

Try setting `FORCE_COLOR=1` environment variable to force enabling color output. Available in v0.23.0+.

---
### How can I version apps deployed by kapp?

kapp itself does not provide any notion of versioning, since it's just a tool to reconcile config. We recommend to include a ConfigMap in your deployment with application metadata e.g. git commit, release notes, etc.

---
#### `Resource ... is associated with a different label value`

Resource ownership is tracked by app labels. kapp expects that each resource is owned by exactly one app.

If you are receiving this error and are using correct app name, it might be that you are targeting wrong namespace where app is located. Use `--namespace` to set correct namespace.

Additional resources: [State Namespace](state-namespace.md), [Slack Thread](https://kubernetes.slack.com/archives/CH8KCCKA5/p1589264289257000)

---
#### Why does kapp hang when trying to delete a resource?

By default, kapp won't delete resources it didn't create. You can see which resources are owned by kapp in output of `kapp inspect -a app-name` in its `Owner` column. You can force kapp to apply this ignored change using `--apply-ignored` [flag](apply.md#controlling-apply-via-deploy-flags). Alternatively if you are able to set [kapp.k14s.io/owned-for-deletion](apply.md#kappk14sioowned-for-deletion) annotation on resource that will be created, kapp will take that as a request to "own it" for deletion. This comes in handy for example with PVCs created by StatefulSet.

---
#### How does kapp handle merging?

kapp explicitly decided against _basic_ 3 way merge, instead allowing the user to specify how to resolve conflicts via rebase rules.

Resources: [merge method](merge-method.md), [rebase rules](config.md#rebaserules)

---
#### Can I force an update for a change that does not produce a diff?

If kapp does not detect changes, it won't perform an update. To force changes every time you can set [`kapp.k14s.io/nonce`](apply.md#kappk14siononce) annotation. That way, every time you deploy the resource will appear to have changes.

---
#### How can I remove decorative headings from kapp inspect output?

Use `--tty=false` flag which will disable decorative output. Example: `kapp inspect --raw --tty=false`.

Additional resources: [tty flag in kapp code](https://github.com/k14s/kapp/blob/3f3e207d7198cdedd6985761ecb0d9616a84e305/pkg/kapp/cmd/ui_flags.go#L20)

---
#### How can I get kapp to skip waiting on some resources?

kapp allows to control waiting behavior per resource via [resource annotations](apply-waiting.md#controlling-waiting-via-resource-annotations). When used, the resource will be applied to the cluster, but will not wait to reconcile.
