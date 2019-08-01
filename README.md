# kapp

- Website: https://get-kapp.io
- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)
- [Docs](docs/README.md) with topics about diff, apply, gitops, config, _blog posts and talks_ etc.
- Install: grab prebuilt binaries from the [Releases page](https://github.com/k14s/kapp/releases).

`kapp` (pronounced: `kap`) CLI encourages Kubernetes users to manage resources in bulk by working with "Kubernetes applications" (sets of resources with the same label). It focuses on resource diffing, labeling, deployment and deletion. Unlike tools like Helm, `kapp` considers YAML templating and management of packages outside of its scope, though it works great with tools that generate Kubernetes configuration.

![](docs/kapp-deploy-screenshot.png)

See [https://get-kapp.io](https://get-kapp.io) for detailed example workflow.

Features:

- Works with standard Kubernetes YAMLs
- Focuses exclusively on deployment workflow, not packaging or templating
  - but plays well with tools (such as [ytt](https://get-ytt.io)) that produce Kubernetes configuration
- Converges application resources (creates, updates and/or deletes resources) in each deploy
  - based on comparison between provided files and live objects in the cluster
- Separates calculation of changes ([diff stage](docs/diff.md)) from application of changes ([apply stage](docs/apply.md))
- [Waits for resources](docs/apply-waiting.md) to be "ready"
- Creates CRDs and Namespaces first and supports [custom change ordering](docs/apply-ordering.md)
- Works [without admin privileges](docs/rbac.md) and does not use custom CRDs
  - making it possible to use kapp as a regular user in a single namespace
- Records application deployment history
- Opt-in resource version management
  - for example, to trigger Deployment rollout when ConfigMap changes
- Works with any group of labeled resources (`kapp -a label:tier=web inspect -t`)
- Works without server side components
- GitOps friendly (`kapp app-group deploy -g all-apps --directory .`)

## Development

```bash
./hack/build.sh
export KAPP_E2E_NAMESPACE=kapp-test
./hack/test-all.sh

# include goog analytics in 'kapp website' command for https://get-kapp.io
# (goog analytics is _not_ included in release binaries)
BUILD_VALUES=./hack/build-values-get-kapp-io.yml ./hack/build.sh
```

`build.sh` depends on [ytt](https://github.com/k14s/ytt).
