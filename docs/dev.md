## Building

```bash
./hack/build.sh
export KAPP_E2E_NAMESPACE=kapp-test
./hack/test-all.sh

# include goog analytics in 'kapp website' command for https://get-kapp.io
# (goog analytics is _not_ included in release binaries)
BUILD_VALUES=./hack/build-values-get-kapp-io.yml ./hack/build.sh
```

`build.sh` depends on [ytt](https://github.com/k14s/ytt).

## Source Code Structure

For those interested in extending and improving `kapp`, below is a quick reference on the structure of the source code:

- [.github/workflows/test-gh.yml](https://github.com/k14s/kapp/blob/develop/.github/workflows/test-gh.yml) is a Github Action that runs build and unit tests when commits are pushed
- [hack](https://github.com/k14s/kapp/tree/develop/hack) has build and test scripts
- [cmd/kapp](https://github.com/k14s/kapp/blob/develop/cmd/kapp) is the entry package for main kapp binary
- [cmd/kapp-lambda-website](https://github.com/k14s/kapp/blob/develop/cmd/kapp-lambda-website) is the entry package for AWS Lambda compatible binary that wraps `kapp website` command
- [pkg/kapp/cmd](https://github.com/k14s/kapp/tree/develop/pkg/kapp/cmd) includes all kapp CLI commands (kapp.go is root command)
  - [pkg/kapp/cmd/app](https://github.com/k14s/kapp/tree/develop/pkg/kapp/cmd/app) includes all top level CLI commands (deploy, delete, etc.)
- [pkg/kapp/app](https://github.com/k14s/kapp/tree/develop/pkg/kapp/app) package includes two types of app and details about them (such as app change tracking):
  - `LabeledApp` which represents app based on user provided label
  - `RecordedApp` which represents app backed by `ConfigMap`. `RecordedApp` uses LabeledApp internally.
- [pkg/kapp/resources](https://github.com/k14s/kapp/tree/develop/pkg/kapp/resources) package is responsible for parsing, fetching, and modifying k8s resources (represented through Resource interface and ResourceImpl)
- [pkg/kapp/diff](https://github.com/k14s/kapp/tree/develop/pkg/kapp/diff) allows to diff two resources represented by Change object, and multiple resources via ChangeSet. **This package does not know anything about k8s.**
- [pkg/kapp/diffgraph](https://github.com/k14s/kapp/tree/develop/pkg/kapp/diffgraph) applies deploy/delete order to set of changes. **This package does not know anything about k8s.**
- [pkg/kapp/clusterapply](https://github.com/k14s/kapp/tree/develop/pkg/kapp/clusterapply) allow to apply diff to k8s cluster via ClusterChange object.
  - [converged_resource.go](https://github.com/k14s/kapp/blob/develop/pkg/kapp/clusterapply/converged_resource.go) tracks whether resource is in a _converged state_ e.g. `Deployment` has finished updating its `Pods`. Uses `resourcesmisc` package as part of its implementation.
  - [add_or_update_change.go](https://github.com/k14s/kapp/blob/develop/pkg/kapp/clusterapply/add_or_update_change.go) controls how resource is created or updated with necessary retry logic
  - [delete_change.go](https://github.com/k14s/kapp/blob/develop/pkg/kapp/clusterapply/delete_change.go) controls how resource is deleted
- [pkg/kapp/resourcesmisc](https://github.com/k14s/kapp/tree/develop/pkg/kapp/resourcesmisc) contains objects for waiting on different resource types (e.g. Deployment, Service, Pod, etc.)
- [pkg/kapp/logs](https://github.com/k14s/kapp/tree/develop/pkg/kapp/logs) supports log streaming for `kapp logs` command
- [test/e2e](https://github.com/k14s/kapp/tree/develop/test/e2e) includes e2e tests that can run against any k8s cluster.
- [pkg/kapp/website](https://github.com/k14s/kapp/tree/develop/pkg/kapp/website) has HTML and JS assets used by `kapp website` command and ultimately https://get-kapp.io.

### Design Principles

- clearly separate diff and apply stages (in code and UI)
  - diffing should not require k8s cluster access beyond initial resource fetching
- all k8s cluster changes must be presented in the UI
  - to make this work as reliably as possible, UI presents ClusterChanges
  - 1 violation: state management of app records and app changes (as ConfigMaps)
- isolate k8s cluster modification to single package (`pkg/kapp/clusterapply`)
- app delete = deploy with no resources + state records deletion
  - i.e. deploy and delete should follow same semantics and code path (`ClusterChangeSet`)
- non-admin users must be able to use kapp against their cluster (single locked down namespace)
