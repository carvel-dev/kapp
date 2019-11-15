## Dangerous Flags

There are several flags in `kapp deploy/delete/etc.` commands that might be helpful in rare cases, but can cause problems if used improperly. These are their stories: 

### `--dangerous-allow-empty-list-of-resources`

This flag allows `kapp deploy` to accept empty set of new resources. Given that kapp deploy converges set of resources, when empty set is provided, kapp will delete all existing resources.

This commonly happens unintentionally. When configuration is piped into kapp (e.g. `ytt -f config/ | kapp deploy -f- ...`) and resource producing command fails (ytt in this example), kapp will not receive any resources by the time is closes. Since providing empty set of resources intentionally is pretty rare, this functionality is behind a flag.

### `--dangerous-override-ownership-of-existing-resources`

This flag allows `kapp deploy` to take ownership of resources that are already associated with another application (i.e. already has `kapp.k14s.io/app` label with a different value).

Most commonly user may have _unintentionally_ included resource that is already deployed, hence by default we do not want to override that resource with a new copy. This may happen when multiple apps accidently specified same resource (i.e. same name under same namespace). In most cases this is not what user wants.

This flag may be useful in cases when multiple applications (managed by kapp) need to be merged into one, or may be previously owning application have been deleted but its resources were kept.

Note that by default if resource is given to kapp and it already exists in the cluster, and is not owned by another application, kapp will label it to belong to deploying app.

### `--dangerous-ignore-failing-api-services`

In some cases users may encounter that they have misbehaving `APIServices` within they cluster. Since `APIServices` affect how one finds existing resources within a cluster, by default kapp will show error similar to below and stop:

```
Error: ... unable to retrieve the complete list of server APIs: <...>: the server is currently unable to handle the request
```

In cases when APIService cannot be fixed, this flag can be used to let kapp know that it is okay to proceed even though it's not able to see resources under that `APIService`. Note when this flag is used, kapp will effectively think that resources under misbehaving `APIService` do not exist.
