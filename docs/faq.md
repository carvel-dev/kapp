## FAQ

### `Error: Asking for confirmation: EOF`

This probably means you have piped configuration into kapp and did not specify `--yes` (`-y`) flag to continue. It's necessary because kapp can no longer ask for confirmation via stdin. Feel free to re-run the command with `--diff-changes` (`-c`) to make sure pending changes are correct.

---
### Where to store app resources (i.e. in which namespace)?

See [state namespace](state-namespace.md) doc page.

---
### `... Field is immutable` error

> After changing the labels/selectors in one of my templates, I'm getting the `MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable (reason: Invalid)` errors on deployment resource. Is there a way to tell kapp to force the change?

[via slack](https://kubernetes.slack.com/archives/CH8KCCKA5/p1565600090224400)

Some fields on a resource are immutable. kapp provides a `kapp.k14s.io/update-strategy` annotation that controls how kapp will update resource. One of the strategies is `fallback-on-replace` which will have kapp recreate an object (delete, wait, then create) if initial update results in `Invalid` error. See [Controlling apply via resource annotations](https://github.com/k14s/kapp/blob/master/docs/apply.md#controlling-apply-via-resource-annotations) for details.

---
### Updating Deployments when ConfigMap changes

> Can kapp force update on ConfigMaps in Deployments/DaemonSets? Just noticed that it didn't do that and I somehow expected it to.

[via slack](https://kubernetes.slack.com/archives/CH8KCCKA5/p1565624685226400)

kapp has a feature called [versioned resources](https://github.com/k14s/kapp/blob/master/docs/diff.md#versioned-resources) that allows kapp to create uniquely named resources instead of updating resources with changes. Resources referencing versioned resources are forced to be updated with new names, and therefore are changed, thus solving a problem of how to propagate changes safely.

---
### Quick way to find common kapp command variations

See [cheatsheet](cheatsheet.md).
