## Merge Method: Why not basic 3 way merge?

kapp explicitly decided to _not_ do basic 3 way merge, and instead allow the user to specify how to resolve "conflicts". Here is our thinking:

- you as an operator have a set of files (input files given to kapp via -f) which describe desired configuration
- cluster has resources that need to be converged to whatever input files specify, with one exception: in some cases, cluster is the source of truth for certain information (but not most) and should keep that state on resources (common examples: some annotation on Deployment, clusterIP on Service, etc.)

Given information above there are multiple ways to converge:

- make assumptions about how to merge things (basic 3 way merge, what kubectl and helm does afaik)
- be explicit about how to merge things (kapp with rebase rules)
- or, just override

Overriding is not really an option as it removes potentially important cluster changes (e.g. removes replicas value as scaled by HPA).

Regarding explicit vs implicit: we decided to go with the explicit option. kapp allows users to add [rebase rules](https://github.com/k14s/kapp/blob/develop/docs/config.md) to specify exactly which information to retain from existing resources. That gives control to the user to decide what's important to be kept based on cluster state and what's not. This method ensures that there are no _surprising_ changes left in the cluster (if basic 3 way merge was used, then user cannot confidently know how final resource will look like; ... imagine if you had a field `allowUnauthenticatedRequests: true` in some resource that someone flipped on in your cluster, and your configs never specified it; it would not be removed unless you decide to also specify this field in your configs).

kapp comes with some common k8s rebase rules. you can see them via `kapp deploy-config`.

tldr: kapp takes user provided config as the only source of truth, but also allows to explicitly specify that certain fields are cluster controlled. This method guarantees that clusters don't drift, which is better than what basic 3 way merge provides.

Originally answered [here](https://github.com/k14s/kapp/issues/58#issuecomment-559214883).
