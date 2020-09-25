## Running kapp under restricted permissions

In a multi-tenant Kubernetes cluster, user's actions may be limited to one or more namespaces via `Role` and `RoleBinding` configuration.

Following setup is currently expected by kapp (v0.10.0+):

- [required] kapp requires list/get/create/update/delete for `v1/ConfigMap` in [state namespace](state-namespace.md) so that it can store record of application and deployment history.
- [optional] kapp requires one `ClusterRole` rule: listing of namespaces. This requirement is necessary for kapp to find all namespaces so that it can search in each namespace resources that belong to a particular app (via a label). As of v0.11.0+, kapp will fallback to only [state namespace](state-namespace.md) if it is forbidden to list all namespaces.
- otherwise, kapp does _not_ require permissions to resource types that are not used in deployed configuration. In other words, if you are not deploying `Job` resource then kapp does not need any permissions for `Job`. Note that some resources are "cluster" created (e.g. `Pods` are created by k8s deployment controller when `Deployment` resource is created) hence users may not see all app associated resources in `kapp inspect` command if they are restricted (this could be advantageous and disadvantegeous in different setups).

Please reach out to us in #carvel channel in k8s slack (linked in [README.md](../README.md)) if current kapp permissions model isn't compatible with your use cases. We are eager to learn about your setup and potentially improve kapp.

Example of `Namespace` listing permission needed by kapp:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kapp-restricted-cr
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kapp-restricted-cr-binding
subjects:
- kind: ServiceAccount
  name: # ???
  namespace: # ??? (some tenant ns)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kapp-restricted-cr
```

Example of `ConfigMap` permissions needed by kapp:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kapp-restricted-role
  namespace: # ??? (some tenant ns)
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["list", "get", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kapp-restricted-role-binding
  namespace: # ??? (some tenant ns)
subjects:
- kind: ServiceAccount
  name: # ???
  namespace: # ??? (some tenant ns)
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kapp-restricted-role
```
